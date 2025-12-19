//go:build e2e

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/stretchr/testify/assert"
)

const (
	e2eDir = "test_e2e"
)

func resetBuildParams() {
	BuildParams = &BuildArgs{}
}

func TestIgnores(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_ignores")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_ignores", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	matches, err := filepath.Glob(filepath.Join(tempNewOutput, "*.zip"))
	assert.NoError(t, err)
	assert.Len(t, matches, 0, "Expected 0 zip files, got %d", len(matches))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestIgnores"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "TestIgnores.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "embed.xml"))
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "ignore_me.old"), "Ignored ignore_me.old file found")
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "ignore_me.new"), "Ignored ignore_me.new file found")
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "example.jpg"), "Ignored example.jpg file found")
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules"))
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "Debug.lua"), "Ignored Debug.lua file found")
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "debug.jpg"), "Ignored debug.jpg file found")
	assert.NoFileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "ignore_me.always"), "Ignored ignore_me.always file found")
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "Suit"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "Suit", "Core.lua"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "Hat"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestIgnores", "Modules", "Hat", "Core.lua"))
}

func TestSvnExternals(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_svn_externals")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_svn_externals", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = true

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "TestSvnExternals.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "embed.xml"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "LibStub"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "LibStub", "LibStub.lua"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "CallbackHandler-1.0"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "CallbackHandler-1.0", "CallbackHandler-1.0.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "CallbackHandler-1.0", "CallbackHandler-1.0.xml"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceAddon-3.0"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceAddon-3.0", "AceAddon-3.0.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceAddon-3.0", "AceAddon-3.0.xml"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceBucket-3.0"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceBucket-3.0", "AceBucket-3.0.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestSvnExternals", "Libs", "AceBucket-3.0", "AceBucket-3.0.xml"))
}

func TestGitExternals(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_git_externals")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_git_externals", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = true

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibClassicSwingTimerAPI"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibClassicSwingTimerAPI", "LibClassicSwingTimerAPI.lua"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibDataBroker-1.1"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibDataBroker-1.1", "LibDataBroker-1.1.lua"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibDeflate"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibDeflate", "LibDeflate.lua"))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibSpellRange-1.0"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestGitExternals", "Libs", "LibSpellRange-1.0", "LibSpellRange-1.0.lua"))
}

func TestZip(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_zip")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_zip", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = false
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	matches, err := filepath.Glob(filepath.Join(tempNewOutput, "*.zip"))
	assert.NoError(t, err)
	assert.Len(t, matches, 1, "Expected 1 zip file, got %d", len(matches))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestZip"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestZip", "TestZip.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestZip", "Core.lua"))
}

func TestZipNoLib(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_zip_nolib")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_zip_nolib", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = false
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	matches, err := filepath.Glob(filepath.Join(tempNewOutput, "*.zip"))
	assert.NoError(t, err)
	assert.Len(t, matches, 2, "Expected 2 zip file(s), got %d", len(matches))
	assert.DirExists(t, filepath.Join(tempNewOutput, "TestZipNoLib"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestZipNoLib", "TestZipNoLib.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestZipNoLib", "Core.lua"))
}

func TestManualChangelog(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_manual_changelog")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_manual_changelog", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestManualChangelog"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestManualChangelog", "TestManualChangelog.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestManualChangelog", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestManualChangelog", "CHANGELOG.txt"))
}

func TestChangelogTitle(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_changelog_title")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_changelog_title", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestChangelogTitle"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestChangelogTitle", "TestChangelogTitle.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestChangelogTitle", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestChangelogTitle", "CHANGELOG.md"))
	contents, err := os.ReadFile(filepath.Join(tempNewOutput, "TestChangelogTitle", "CHANGELOG.md"))
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "TEST CHANGELOG TITLE")
}

func TestLicenseExist(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_license_exist")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_license_exist", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = false

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestLicenseExist"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseExist", "TestLicenseExist.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseExist", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseExist", "my_license.txt"))
	contents, err := os.ReadFile(filepath.Join(tempNewOutput, "TestLicenseExist", "my_license.txt"))
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "License to view")
}

func TestLicenseDownload(t *testing.T) {
	defer resetBuildParams()
	testDir := filepath.Join(".", e2eDir, "test_license_download")
	tempNewOutput, err := filepath.Abs(filepath.Join(".", e2eDir, "test_license_download", ".release"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(tempNewOutput); err == nil {
		if err := os.RemoveAll(tempNewOutput); err != nil {
			t.Fatalf("Failed to remove old output directory: %v", err)
		}
	}

	BuildParams.SkipUpload = true
	BuildParams.SkipZip = true
	BuildParams.ForceExternals = false
	BuildParams.CurseId = "1082791"

	runNewCLI(t, testDir, tempNewOutput)

	assert.DirExists(t, filepath.Join(tempNewOutput, "TestLicenseDownload"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseDownload", "TestLicenseDownload.toc"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseDownload", "Core.lua"))
	assert.FileExists(t, filepath.Join(tempNewOutput, "TestLicenseDownload", "my_license.txt"))
	contents, err := os.ReadFile(filepath.Join(tempNewOutput, "TestLicenseDownload", "my_license.txt"))
	strContents := string(contents)
	assert.NoError(t, err)
	assert.NotEmpty(t, strContents)
	assert.Contains(t, strContents, "MIT")
	assert.Contains(t, strContents, strconv.Itoa(time.Now().UTC().Year()))
	assert.Contains(t, strContents, "McTalian")
	assert.NotContains(t, strContents, "<p>")
}

func runNewCLI(t *testing.T, input, output string) {
	// Capture stdout/stderr if needed
	oldStdout, oldStderr := os.Stdout, os.Stderr
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr }() // Restore after execution

	_, wOut, _ := os.Pipe()
	_, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	logger.InitLogger()
	BuildParams.TopDir = input
	BuildParams.ReleaseDir = output
	err := Build(BuildParams)
	if err != nil {
		assert.NoError(t, fmt.Errorf("failed to run new CLI: %v", err))
		t.FailNow()
	}

	// Close the write ends of the pipes
	err = wOut.Close()
	assert.NoError(t, err)
	err = wErr.Close()
	assert.NoError(t, err)
}
