package github

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed release_schema.json
var schemaFS embed.FS

var (
	releaseSchema     *jsonschema.Schema
	releaseSchemaOnce sync.Once
	releaseSchemaErr  error
)

func getReleaseSchema() (*jsonschema.Schema, error) {
	releaseSchemaOnce.Do(func() {
		schemaRaw, err := schemaFS.ReadFile("release_schema.json")
		if err != nil {
			releaseSchemaErr = fmt.Errorf("failed to read release schema: %w", err)
			return
		}

		var schemaDoc any
		if err := json.Unmarshal(schemaRaw, &schemaDoc); err != nil {
			releaseSchemaErr = fmt.Errorf("failed to unmarshal release schema: %w", err)
			return
		}

		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource("release_schema.json", schemaDoc); err != nil {
			releaseSchemaErr = fmt.Errorf("failed to add schema resource: %w", err)
			return
		}

		schema, err := compiler.Compile("release_schema.json")
		if err != nil {
			releaseSchemaErr = fmt.Errorf("failed to compile release schema: %w", err)
			return
		}
		releaseSchema = schema
	})

	return releaseSchema, releaseSchemaErr
}

func ValidateReleaseJSON(releaseJSON string) error {
	schema, err := getReleaseSchema()
	if err != nil {
		return fmt.Errorf("failed to get release schema: %w", err)
	}

	var doc any
	if err := json.Unmarshal([]byte(releaseJSON), &doc); err != nil {
		return fmt.Errorf("failed to unmarshal release json for validation: %w", err)
	}

	if err := schema.Validate(doc); err != nil {
		return fmt.Errorf("release.json validation failed: %w", err)
	}

	return nil
}

type metadata struct {
	Flavor    string `json:"flavor"`
	Interface int    `json:"interface"`
}

type wbtRelease struct {
	Name     string     `json:"name"`
	Version  string     `json:"version"`
	Filename string     `json:"filename"`
	NoLib    bool       `json:"nolib"`
	Metadata []metadata `json:"metadata"`
}

type wbtReleaseMetadata struct {
	Releases []wbtRelease `json:"releases"`
}

func GetReleaseMetadataContents(name string, version string, gameInterfaces toc.GameInterfaces, zipFileNames ...string) (string, error) {
	releaseMetadata := wbtReleaseMetadata{
		Releases: make([]wbtRelease, 0),
	}
	for _, zipFileName := range zipFileNames {
		noLib := strings.Contains(zipFileName, "nolib")
		release := wbtRelease{
			Name:     name,
			Version:  version,
			Filename: filepath.Base(zipFileName),
			NoLib:    noLib,
		}

		for flavor, versions := range gameInterfaces {
			for _, v := range versions {
				flavorStr := flavor.ToString()
				if flavor == toc.Retail {
					flavorStr = "mainline"
				}
				release.Metadata = append(release.Metadata, metadata{
					Flavor:    flavorStr,
					Interface: v,
				})
			}
		}
		releaseMetadata.Releases = append(releaseMetadata.Releases, release)
	}

	contents, err := json.Marshal(&releaseMetadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal release metadata: %w", err)
	}

	return string(contents), nil
}
