package github

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const remoteSchemaURL = "https://raw.githubusercontent.com/ogri-la/release.json-specification/refs/heads/master/schema.json"

func TestEmbeddedSchemaMatchesRemote(t *testing.T) {
	resp, err := http.Get(remoteSchemaURL)
	if err != nil {
		t.Logf("Could not fetch remote schema (%v), skipping", err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Logf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Logf("Remote schema returned status %d, skipping", resp.StatusCode)
		return
	}

	var remoteDoc any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&remoteDoc))

	localRaw, err := schemaFS.ReadFile("release_schema.json")
	require.NoError(t, err)

	var localDoc any
	require.NoError(t, json.Unmarshal(localRaw, &localDoc))

	remoteJSON, _ := json.Marshal(remoteDoc)
	localJSON, _ := json.Marshal(localDoc)

	assert.Equal(t, string(localJSON), string(remoteJSON),
		"Embedded schema differs from remote. Run 'curl -o internal/github/release_schema.json %s' to update.", remoteSchemaURL)
}

func TestGetReleaseMetadataContents(t *testing.T) {
	tests := []struct {
		name             string
		packageName      string
		version          string
		gameInterfaces   toc.GameInterfaces
		zipFileNames     []string
		expectedReleases int
	}{
		{
			name:        "single zip file with retail",
			packageName: "MyAddon",
			version:     "1.0.0",
			gameInterfaces: toc.GameInterfaces{
				toc.Retail: []int{110000, 110002},
			},
			zipFileNames:     []string{"MyAddon-1.0.0.zip"},
			expectedReleases: 1,
		},
		{
			name:        "multiple zip files",
			packageName: "MyAddon",
			version:     "1.0.0",
			gameInterfaces: toc.GameInterfaces{
				toc.Retail:         []int{110000},
				toc.CurrentClassic: []int{11503},
			},
			zipFileNames:     []string{"MyAddon-1.0.0.zip", "MyAddon-1.0.0-nolib.zip"},
			expectedReleases: 2,
		},
		{
			name:        "nolib zip file",
			packageName: "MyAddon",
			version:     "2.0.0",
			gameInterfaces: toc.GameInterfaces{
				toc.Retail: []int{110000},
			},
			zipFileNames:     []string{"MyAddon-2.0.0-nolib.zip"},
			expectedReleases: 1,
		},
		{
			name:        "multiple flavors",
			packageName: "MyAddon",
			version:     "1.5.0",
			gameInterfaces: toc.GameInterfaces{
				toc.Retail:         []int{110000},
				toc.CurrentClassic: []int{11503},
				toc.ClassicEra:     []int{11502},
			},
			zipFileNames:     []string{"MyAddon-1.5.0.zip"},
			expectedReleases: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetReleaseMetadataContents(tt.packageName, tt.version, tt.gameInterfaces, tt.zipFileNames...)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			// Parse the JSON to verify structure
			var metadata wbtReleaseMetadata
			err = json.Unmarshal([]byte(result), &metadata)
			require.NoError(t, err)

			assert.Len(t, metadata.Releases, tt.expectedReleases)

			for _, release := range metadata.Releases {
				assert.Equal(t, tt.packageName, release.Name)
				assert.Equal(t, tt.version, release.Version)

				// Check filename is just the base name
				assert.NotContains(t, release.Filename, "/")
				assert.NotContains(t, release.Filename, "\\")

				// Check metadata entries exist
				assert.NotEmpty(t, release.Metadata, "release should have metadata")
			}
		})
	}
}

func TestGetReleaseMetadataContents_FlavorMapping(t *testing.T) {
	packageName := "TestAddon"
	version := "1.0.0"
	gameInterfaces := toc.GameInterfaces{
		toc.Retail:         []int{110000},
		toc.TbcClassic:     []int{20500},
		toc.CurrentClassic: []int{11503},
	}
	zipFileNames := []string{"TestAddon-1.0.0.zip"}

	result, err := GetReleaseMetadataContents(packageName, version, gameInterfaces, zipFileNames...)
	require.NoError(t, err)

	var metadata wbtReleaseMetadata
	err = json.Unmarshal([]byte(result), &metadata)
	require.NoError(t, err)

	require.Len(t, metadata.Releases, 1)
	release := metadata.Releases[0]

	// Check that retail is mapped to "mainline"
	hasMainline := false
	hasBcc := false
	hasMists := false
	for _, m := range release.Metadata {
		if m.Flavor == "mainline" {
			hasMainline = true
			assert.Equal(t, 110000, m.Interface)
		}
		if m.Flavor == "bcc" {
			hasBcc = true
			assert.Equal(t, 20500, m.Interface)
		}
		if m.Flavor == "mists" {
			hasMists = true
			assert.Equal(t, 11503, m.Interface)
		}
	}

	assert.True(t, hasMainline, "retail should be mapped to mainline")
	assert.True(t, hasBcc, "burning crusade classic should be mapped to bcc")
	assert.True(t, hasMists, "current classic (mists) should be present")
}

func TestGetReleaseMetadataContents_NoLibDetection(t *testing.T) {
	packageName := "TestAddon"
	version := "1.0.0"
	gameInterfaces := toc.GameInterfaces{
		toc.Retail: []int{110000},
	}

	tests := []struct {
		name          string
		zipFileName   string
		expectedNoLib bool
	}{
		{
			name:          "regular zip",
			zipFileName:   "TestAddon-1.0.0.zip",
			expectedNoLib: false,
		},
		{
			name:          "nolib zip",
			zipFileName:   "TestAddon-1.0.0-nolib.zip",
			expectedNoLib: true,
		},
		{
			name:          "NoLib with capital letters",
			zipFileName:   "TestAddon-1.0.0-NoLib.zip",
			expectedNoLib: false, // Case sensitive check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetReleaseMetadataContents(packageName, version, gameInterfaces, tt.zipFileName)
			require.NoError(t, err)

			var metadata wbtReleaseMetadata
			err = json.Unmarshal([]byte(result), &metadata)
			require.NoError(t, err)

			require.Len(t, metadata.Releases, 1)
			assert.Equal(t, tt.expectedNoLib, metadata.Releases[0].NoLib)
		})
	}
}

func TestGetReleaseMetadataContents_EmptyInputs(t *testing.T) {
	result, err := GetReleaseMetadataContents("", "", toc.GameInterfaces{})
	require.NoError(t, err)

	var metadata wbtReleaseMetadata
	err = json.Unmarshal([]byte(result), &metadata)
	require.NoError(t, err)

	assert.Empty(t, metadata.Releases)
}

func TestValidateReleaseJSON_Valid(t *testing.T) {
	valid := `{
		"releases": [
			{
				"name": "MyAddon",
				"version": "1.0.0",
				"filename": "MyAddon-1.0.0.zip",
				"nolib": false,
				"metadata": [
					{ "flavor": "mainline", "interface": 110000 }
				]
			}
		]
	}`
	err := ValidateReleaseJSON(valid)
	assert.NoError(t, err)
}

func TestValidateReleaseJSON_InvalidFlavor(t *testing.T) {
	invalid := `{
		"releases": [
			{
				"name": "MyAddon",
				"version": "1.0.0",
				"filename": "MyAddon-1.0.0.zip",
				"nolib": false,
				"metadata": [
					{ "flavor": "invalid_flavor", "interface": 110000 }
				]
			}
		]
	}`
	err := ValidateReleaseJSON(invalid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}

func TestValidateReleaseJSON_MissingRequiredField(t *testing.T) {
	invalid := `{
		"releases": [
			{
				"name": "MyAddon",
				"version": "1.0.0",
				"filename": "MyAddon-1.0.0.zip",
				"nolib": false,
				"metadata": [
					{ "flavor": "mainline" }
				]
			}
		]
	}`
	err := ValidateReleaseJSON(invalid)
	assert.Error(t, err)
}

func TestValidateReleaseJSON_MissingReleases(t *testing.T) {
	invalid := `{}`
	err := ValidateReleaseJSON(invalid)
	assert.Error(t, err)
}

func TestValidateReleaseJSON_AdditionalProperty(t *testing.T) {
	invalid := `{
		"releases": [
			{
				"name": "MyAddon",
				"version": "1.0.0",
				"filename": "MyAddon-1.0.0.zip",
				"nolib": false,
				"extraField": "not allowed",
				"metadata": [
					{ "flavor": "mainline", "interface": 110000 }
				]
			}
		]
	}`
	err := ValidateReleaseJSON(invalid)
	assert.Error(t, err)
}
