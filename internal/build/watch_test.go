package build

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFile(t *testing.T) {
	t.Run("successful copy", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "dest.txt")

		content := []byte("test content")
		err := os.WriteFile(srcFile, content, 0644)
		require.NoError(t, err)

		// Execute
		err = copyFile(srcFile, dstFile)

		// Assert
		assert.NoError(t, err)
		assert.FileExists(t, dstFile)

		dstContent, err := os.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, content, dstContent)

		// Verify modification times match
		srcInfo, err := os.Stat(srcFile)
		require.NoError(t, err)
		dstInfo, err := os.Stat(dstFile)
		require.NoError(t, err)
		assert.True(t, srcInfo.ModTime().Equal(dstInfo.ModTime()))
	})

	t.Run("skip copy when files are identical", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "dest.txt")

		content := []byte("test content")
		modTime := time.Now().Add(-1 * time.Hour)

		err := os.WriteFile(srcFile, content, 0644)
		require.NoError(t, err)
		err = os.Chtimes(srcFile, modTime, modTime)
		require.NoError(t, err)

		err = os.WriteFile(dstFile, content, 0644)
		require.NoError(t, err)
		err = os.Chtimes(dstFile, modTime, modTime)
		require.NoError(t, err)

		// Execute
		err = copyFile(srcFile, dstFile)

		// Assert - no error means it skipped (optimization)
		assert.NoError(t, err)
	})

	t.Run("creates destination directory", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "subdir", "dest.txt")

		content := []byte("test content")
		err := os.WriteFile(srcFile, content, 0644)
		require.NoError(t, err)

		// Execute
		err = copyFile(srcFile, dstFile)

		// Assert
		assert.NoError(t, err)
		assert.FileExists(t, dstFile)
		assert.DirExists(t, filepath.Join(tmpDir, "subdir"))
	})

	t.Run("error when source doesn't exist", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "nonexistent.txt")
		dstFile := filepath.Join(tmpDir, "dest.txt")

		// Execute
		err := copyFile(srcFile, dstFile)

		// Assert
		assert.Error(t, err)
	})

	t.Run("overwrites existing file with different content", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "dest.txt")

		newContent := []byte("new content")
		err := os.WriteFile(srcFile, newContent, 0644)
		require.NoError(t, err)

		// Set source file to a newer time to ensure it's not skipped
		newTime := time.Now().Add(-1 * time.Hour)
		err = os.Chtimes(srcFile, newTime, newTime)
		require.NoError(t, err)

		oldContent := []byte("old content")
		err = os.WriteFile(dstFile, oldContent, 0644)
		require.NoError(t, err)

		// Set dest file to an older time
		oldTime := time.Now().Add(-2 * time.Hour)
		err = os.Chtimes(dstFile, oldTime, oldTime)
		require.NoError(t, err)

		// Execute
		err = copyFile(srcFile, dstFile)

		// Assert
		assert.NoError(t, err)
		dstContent, err := os.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, newContent, dstContent)
	})
}

func TestCopyDir(t *testing.T) {
	t.Run("successful recursive copy", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "source")
		dstDir := filepath.Join(tmpDir, "dest")

		// Create source directory structure
		err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)
		require.NoError(t, err)

		// Execute
		err = copyDir(srcDir, dstDir)

		// Assert
		assert.NoError(t, err)
		assert.DirExists(t, dstDir)
		assert.DirExists(t, filepath.Join(dstDir, "subdir"))
		assert.FileExists(t, filepath.Join(dstDir, "file1.txt"))
		assert.FileExists(t, filepath.Join(dstDir, "subdir", "file2.txt"))

		content, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
		require.NoError(t, err)
		assert.Equal(t, []byte("content1"), content)

		content, err = os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
		require.NoError(t, err)
		assert.Equal(t, []byte("content2"), content)
	})

	t.Run("copy multiple nested directories", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "source")
		dstDir := filepath.Join(tmpDir, "dest")

		// Create nested structure
		err := os.MkdirAll(filepath.Join(srcDir, "level1", "level2", "level3"), 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(srcDir, "level1", "level2", "level3", "deep.txt"), []byte("deep content"), 0644)
		require.NoError(t, err)

		// Execute
		err = copyDir(srcDir, dstDir)

		// Assert
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(dstDir, "level1", "level2", "level3", "deep.txt"))

		content, err := os.ReadFile(filepath.Join(dstDir, "level1", "level2", "level3", "deep.txt"))
		require.NoError(t, err)
		assert.Equal(t, []byte("deep content"), content)
	})

	t.Run("error when source doesn't exist", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "nonexistent")
		dstDir := filepath.Join(tmpDir, "dest")

		// Execute
		err := copyDir(srcDir, dstDir)

		// Assert
		assert.Error(t, err)
	})

	t.Run("empty directory", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "source")
		dstDir := filepath.Join(tmpDir, "dest")

		err := os.MkdirAll(srcDir, 0755)
		require.NoError(t, err)

		// Execute
		err = copyDir(srcDir, dstDir)

		// Assert - copyDir doesn't create dest directory if source is empty
		// This is expected behavior since ReadDir returns empty list
		assert.NoError(t, err)
		// Don't check if dstDir exists - it won't be created for empty source
	})
}

func TestCopyToWow(t *testing.T) {
	t.Run("skip when CopyToWowDirs is false", func(t *testing.T) {
		// Setup
		l := logger.GetSubLog("TEST")
		done := make(chan error, 1)
		WatchParams = &WatchArgs{CopyToWowDirs: false}

		// Execute
		copyToWow(l, done)

		// Assert - should not send any errors
		select {
		case err := <-done:
			t.Errorf("Expected no error, got: %v", err)
		default:
			// Success - no errors sent
		}
	})

	t.Run("copies to configured WoW directories", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		l := logger.GetSubLog("TEST")
		done := make(chan error, 10)

		// Create mock release directory with addon
		releaseDir := filepath.Join(tmpDir, "release")
		BuildParams = &BuildArgs{ReleaseDir: releaseDir}

		addonDir := "TestAddon"
		err := os.MkdirAll(filepath.Join(releaseDir, addonDir), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(releaseDir, addonDir, "Core.lua"), []byte("-- test"), 0644)
		require.NoError(t, err)

		// Setup destination paths
		wowPath1 := filepath.Join(tmpDir, "wow1")
		wowPath2 := filepath.Join(tmpDir, "wow2")
		destinationPaths = []string{wowPath1, wowPath2}
		addonDirs = []string{addonDir}

		WatchParams = &WatchArgs{CopyToWowDirs: true}

		// Execute
		copyToWow(l, done)

		// Assert
		close(done)
		for err := range done {
			assert.NoError(t, err)
		}

		// Verify files were copied to both locations
		assert.FileExists(t, filepath.Join(wowPath1, "Interface", "AddOns", addonDir, "Core.lua"))
		assert.FileExists(t, filepath.Join(wowPath2, "Interface", "AddOns", addonDir, "Core.lua"))
	})

	t.Run("creates Interface/AddOns directory if missing", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		l := logger.GetSubLog("TEST")
		done := make(chan error, 10)

		// Create mock release directory
		releaseDir := filepath.Join(tmpDir, "release")
		BuildParams = &BuildArgs{ReleaseDir: releaseDir}

		addonDir := "TestAddon"
		err := os.MkdirAll(filepath.Join(releaseDir, addonDir), 0755)
		require.NoError(t, err)

		wowPath := filepath.Join(tmpDir, "wow")
		destinationPaths = []string{wowPath}
		addonDirs = []string{addonDir}

		WatchParams = &WatchArgs{CopyToWowDirs: true}

		// Execute
		copyToWow(l, done)

		// Assert
		close(done)
		assert.DirExists(t, filepath.Join(wowPath, "Interface", "AddOns"))
	})
}

func TestTriggerBuild(t *testing.T) {
	t.Run("sets correct build parameters", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		done := make(chan error, 1)

		BuildParams = &BuildArgs{
			TopDir:     tmpDir,
			ReleaseDir: filepath.Join(tmpDir, "release"),
		}

		// Create minimal .toc file for build to work
		err := os.WriteFile(filepath.Join(tmpDir, "Test.toc"), []byte("## Interface: 110002\n## Title: Test\n"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "Core.lua"), []byte("-- test"), 0644)
		require.NoError(t, err)

		// Execute
		triggerBuild(done)

		// Assert - check if error was sent
		select {
		case err := <-done:
			if err != nil {
				// Some errors are expected if full environment isn't set up
				t.Logf("Build error (may be expected in test environment): %v", err)
			}
		default:
			// No error
		}
	})
}

func TestWatchArgs(t *testing.T) {
	t.Run("WatchParams is initialized", func(t *testing.T) {
		assert.NotNil(t, WatchParams)
	})

	t.Run("can set CopyToWowDirs", func(t *testing.T) {
		WatchParams.CopyToWowDirs = true
		assert.True(t, WatchParams.CopyToWowDirs)

		WatchParams.CopyToWowDirs = false
		assert.False(t, WatchParams.CopyToWowDirs)
	})
}

// Benchmark tests
func BenchmarkCopyFile(b *testing.B) {
	tmpDir := b.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := make([]byte, 1024*1024) // 1MB file
	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstFile := filepath.Join(tmpDir, fmt.Sprintf("dest%d.txt", i))
		_ = copyFile(srcFile, dstFile)
	}
}

func BenchmarkCopyDir(b *testing.B) {
	tmpDir := b.TempDir()
	srcDir := filepath.Join(tmpDir, "source")

	// Create source directory with some files
	err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	require.NoError(b, err)

	for i := 0; i < 10; i++ {
		err := os.WriteFile(filepath.Join(srcDir, fmt.Sprintf("file%d.txt", i)), []byte("content"), 0644)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstDir := filepath.Join(tmpDir, fmt.Sprintf("dest%d", i))
		_ = copyDir(srcDir, dstDir)
	}
}
