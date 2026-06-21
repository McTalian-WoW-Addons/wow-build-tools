package github

import (
	"encoding/json"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
