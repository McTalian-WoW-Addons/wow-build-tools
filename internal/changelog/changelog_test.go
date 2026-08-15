package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestVerifyManualChangelog(t *testing.T) {
	tests := []struct {
		name                string
		preExistingFilePath string
		markupType          MarkupType
		setup               func() (string, string)
		expectedError       error
	}{
		{
			name:                "NoPreExistingFilePath",
			preExistingFilePath: "",
			markupType:          MarkdownMT,
			setup:               func() (string, string) { return "", "" },
			expectedError:       ErrManualChangelogNotFound,
		},
		{
			name:                "FileNotFoundInTopDir",
			preExistingFilePath: "CHANGELOG.md",
			markupType:          MarkdownMT,
			setup: func() (string, string) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				return topDir, pkgDir
			},
			expectedError: ErrManualChangelogNotFound,
		},
		{
			name:                "FileFoundInTopDir",
			preExistingFilePath: "CHANGELOG.md",
			markupType:          MarkdownMT,
			setup: func() (string, string) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				filePath := filepath.Join(topDir, "CHANGELOG.md")
				err := os.WriteFile(filePath, []byte("changelog content"), 0644)
				require.NoError(t, err)
				return topDir, pkgDir
			},
			expectedError: nil,
		},
		{
			name:                "FileFoundInPkgDir",
			preExistingFilePath: "CHANGELOG.md",
			markupType:          MarkdownMT,
			setup: func() (string, string) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				topDirFilePath := filepath.Join(topDir, "CHANGELOG.md")
				err := os.WriteFile(topDirFilePath, []byte("topdir changelog content"), 0644)
				require.NoError(t, err)
				filePath := filepath.Join(pkgDir, "CHANGELOG.md")
				err = os.WriteFile(filePath, []byte("changelog content"), 0644)
				require.NoError(t, err)
				return topDir, pkgDir
			},
			expectedError: nil,
		},
		{
			name:                "InvalidMarkupType",
			preExistingFilePath: "CHANGELOG.md",
			markupType:          "invalid",
			setup: func() (string, string) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				filePath := filepath.Join(topDir, "CHANGELOG.md")
				err := os.WriteFile(filePath, []byte("changelog content"), 0644)
				require.NoError(t, err)
				return topDir, pkgDir
			},
			expectedError: ErrInvalidMarkupType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topDir, pkgDir := tt.setup()
			changelog := &Changelog{
				topDir:              topDir,
				pkgDir:              pkgDir,
				PreExistingFilePath: tt.preExistingFilePath,
				MarkupType:          tt.markupType,
			}
			err := changelog.verifyManualChangelog()
			if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestGetChangelog(t *testing.T) {
	tests := []struct {
		name                string
		preExistingFilePath string
		markupType          MarkupType
		generateChangelog   bool
		setup               func() (string, string, repo.VcsRepo)
		expectedError       error
	}{
		{
			name:                "ManualChangelogExistsInPkgDir",
			preExistingFilePath: "{pkgDir}/CHANGELOG.md",
			markupType:          MarkdownMT,
			generateChangelog:   false,
			setup: func() (string, string, repo.VcsRepo) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				filePath := filepath.Join(pkgDir, "CHANGELOG.md")
				err := os.WriteFile(filePath, []byte("changelog content"), 0644)
				require.NoError(t, err)
				return topDir, pkgDir, nil
			},
			expectedError: nil,
		},
		{
			name:                "ManualChangelogExistsInTopDir",
			preExistingFilePath: "{topDir}/CHANGELOG.md",
			markupType:          MarkdownMT,
			generateChangelog:   false,
			setup: func() (string, string, repo.VcsRepo) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				filePath := filepath.Join(topDir, "CHANGELOG.md")
				err := os.WriteFile(filePath, []byte("changelog content"), 0644)
				require.NoError(t, err)
				return topDir, pkgDir, nil
			},
			expectedError: nil,
		},
		{
			name:                "ManualChangelogNotFound",
			preExistingFilePath: "CHANGELOG.md",
			markupType:          MarkdownMT,
			generateChangelog:   false,
			setup: func() (string, string, repo.VcsRepo) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				mockRepo := &repo.MockVcsRepo{
					GetChangelogFunc: func(title string) (string, error) {
						return "generated changelog contents", nil
					},
				}
				return topDir, pkgDir, mockRepo
			},
			expectedError: nil,
		},
		{
			name:                "GenerateChangelog",
			preExistingFilePath: "",
			markupType:          MarkdownMT,
			generateChangelog:   true,
			setup: func() (string, string, repo.VcsRepo) {
				topDir := t.TempDir()
				pkgDir := t.TempDir()
				mockRepo := &repo.MockVcsRepo{
					GetChangelogFunc: func(title string) (string, error) {
						return "generated changelog contents", nil
					},
				}
				return topDir, pkgDir, mockRepo
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topDir, pkgDir, mockRepo := tt.setup()
			filePath := strings.ReplaceAll(tt.preExistingFilePath, "{topDir}", topDir)
			filePath = strings.ReplaceAll(filePath, "{pkgDir}", pkgDir)
			changelog := &Changelog{
				topDir:              topDir,
				pkgDir:              pkgDir,
				PreExistingFilePath: filePath,
				MarkupType:          tt.markupType,
				generateChangelog:   tt.generateChangelog,
				repo:                mockRepo,
			}
			err := changelog.GetChangelog()
			if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.generateChangelog {
				expectedPath := filepath.Join(pkgDir, "CHANGELOG.md")
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("expected changelog file to be generated at %s", expectedPath)
				}
			}
		})
	}
}

// TestGetChangelog_ReusedPkgDirTruncatesStaleContent locks in a fix for a bug
// where the generated CHANGELOG.md was opened with O_CREATE|O_WRONLY but no
// O_TRUNC. --keepPackageDir (which `make watch` always sets, see
// internal/build/watch.go) reuses pkgDir across builds without wiping it
// first, so if a regenerated changelog is shorter than what was written on a
// previous run, stale trailing bytes from that previous run survived into
// the new CHANGELOG.md instead of being overwritten.
func TestGetChangelog_ReusedPkgDirTruncatesStaleContent(t *testing.T) {
	topDir := t.TempDir()
	pkgDir := t.TempDir()

	first := &Changelog{
		topDir:            topDir,
		pkgDir:            pkgDir,
		MarkupType:        MarkdownMT,
		generateChangelog: true,
		repo: &repo.MockVcsRepo{
			GetChangelogFunc: func(title string) (string, error) {
				return "a much longer changelog entry from the first build", nil
			},
		},
	}
	require.NoError(t, first.GetChangelog())

	second := &Changelog{
		topDir:            topDir,
		pkgDir:            pkgDir,
		MarkupType:        MarkdownMT,
		generateChangelog: true,
		repo: &repo.MockVcsRepo{
			GetChangelogFunc: func(title string) (string, error) {
				return "short", nil
			},
		},
	}
	require.NoError(t, second.GetChangelog())

	contents, err := os.ReadFile(filepath.Join(pkgDir, "CHANGELOG.md"))
	require.NoError(t, err)
	require.Equal(t, "short", string(contents), "stale bytes from the previous, longer changelog should not survive")
}
