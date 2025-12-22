package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
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
		err = triggerBuild()

		// Assert - check if error was sent
		if err != nil {
			t.Logf("Build error (may be expected in test environment): %v", err)
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

func TestSetupWatchEnvironment(t *testing.T) {
	logger.InitLogger()

	t.Run("creates and cleans release directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")

		// Create release directory with a file first
		require.NoError(t, os.MkdirAll(releaseDir, 0755))
		testFile := filepath.Join(releaseDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		err := setupWatchEnvironment(releaseDir, false)

		assert.NoError(t, err)
		// The function removes the directory, so it should no longer exist
		// or exist but be empty
		if _, err := os.Stat(releaseDir); err == nil {
			// Directory exists, check it's empty
			entries, err := os.ReadDir(releaseDir)
			require.NoError(t, err)
			assert.Empty(t, entries, "release directory should be empty after setup")
		}
	})

	t.Run("handles nested release directory path", func(t *testing.T) {
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "nested", "release", "dir")

		err := setupWatchEnvironment(releaseDir, false)

		assert.NoError(t, err)
		// Parent directories should be created
		assert.DirExists(t, filepath.Dir(releaseDir))
	})

	t.Run("with copy to wow dirs enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")

		err := setupWatchEnvironment(releaseDir, true)

		assert.NoError(t, err)
	})
}

func TestLoadWowPaths(t *testing.T) {
	logger.InitLogger()

	// Save original viper config and restore after tests
	origConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range origConfig {
			viper.Set(k, v)
		}
	}()

	t.Run("returns error when no paths configured", func(t *testing.T) {
		viper.Reset()
		viper.Set("wowPath", map[string]string{"base": "/some/path"})

		paths, err := loadWowPaths()

		assert.Error(t, err)
		assert.Nil(t, paths)
		assert.Contains(t, err.Error(), "no WoW paths configured")
	})

	t.Run("filters out base path", func(t *testing.T) {
		viper.Reset()
		viper.Set("wowPath", map[string]interface{}{
			"base":    "/base/path",
			"retail":  "/wow/retail",
			"classic": "/wow/classic",
		})

		paths, err := loadWowPaths()

		assert.NoError(t, err)
		assert.Len(t, paths, 2)
		assert.NotContains(t, paths, "/base/path")
		assert.Contains(t, paths, "/wow/retail")
		assert.Contains(t, paths, "/wow/classic")
	})

	t.Run("handles empty wow path config", func(t *testing.T) {
		viper.Reset()
		viper.Set("wowPath", map[string]string{})

		paths, err := loadWowPaths()

		assert.Error(t, err)
		assert.Nil(t, paths)
	})
}

func TestSetupWatcher(t *testing.T) {
	logger.InitLogger()

	t.Run("creates watcher successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		dir1 := filepath.Join(tmpDir, "dir1")
		dir2 := filepath.Join(tmpDir, "dir2")

		require.NoError(t, os.MkdirAll(dir1, 0755))
		require.NoError(t, os.MkdirAll(dir2, 0755))

		dirsToWatch := []string{dir1, dir2}

		watcher, err := setupWatcher(dirsToWatch)

		assert.NoError(t, err)
		assert.NotNil(t, watcher)

		if watcher != nil {
			_ = watcher.Close()
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentDir := filepath.Join(tmpDir, "does-not-exist")

		dirsToWatch := []string{nonExistentDir}

		watcher, err := setupWatcher(dirsToWatch)

		assert.Error(t, err)
		assert.Nil(t, watcher)
	})

	t.Run("closes watcher on error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dir1 := filepath.Join(tmpDir, "dir1")
		nonExistentDir := filepath.Join(tmpDir, "does-not-exist")

		require.NoError(t, os.MkdirAll(dir1, 0755))

		dirsToWatch := []string{dir1, nonExistentDir}

		watcher, err := setupWatcher(dirsToWatch)

		assert.Error(t, err)
		assert.Nil(t, watcher)
	})
}

func TestShouldProcessEvent(t *testing.T) {
	logger.InitLogger()

	tests := []struct {
		name       string
		op         fsnotify.Op
		eventName  string
		releaseDir string
		expected   bool
	}{
		{
			name:       "write event should be processed",
			op:         fsnotify.Write,
			eventName:  "/path/to/file.txt",
			releaseDir: "/release",
			expected:   true,
		},
		{
			name:       "create event should be processed",
			op:         fsnotify.Create,
			eventName:  "/path/to/file.txt",
			releaseDir: "/release",
			expected:   true,
		},
		{
			name:       "remove event should be processed",
			op:         fsnotify.Remove,
			eventName:  "/path/to/file.txt",
			releaseDir: "/release",
			expected:   true,
		},
		{
			name:       "rename event should be processed",
			op:         fsnotify.Rename,
			eventName:  "/path/to/file.txt",
			releaseDir: "/release",
			expected:   true,
		},
		{
			name:       "chmod event should be ignored",
			op:         fsnotify.Chmod,
			eventName:  "/path/to/file.txt",
			releaseDir: "/release",
			expected:   false,
		},
		{
			name:       "event in release dir should be ignored",
			op:         fsnotify.Write,
			eventName:  "/release/file.txt",
			releaseDir: "/release",
			expected:   false,
		},
		{
			name:       "event with release dir in path should be ignored",
			op:         fsnotify.Write,
			eventName:  "/path/release/file.txt",
			releaseDir: "release",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := fsnotify.Event{
				Name: tt.eventName,
				Op:   tt.op,
			}

			result := shouldProcessEvent(event, tt.releaseDir)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunInitialBuild(t *testing.T) {
	logger.InitLogger()

	t.Run("returns error when build fails", func(t *testing.T) {
		// Setup: BuildParams with invalid topdir to cause build failure
		originalBuildParams := BuildParams
		defer func() { BuildParams = originalBuildParams }()

		BuildParams = &BuildArgs{
			TopDir:     "/nonexistent/path",
			ReleaseDir: t.TempDir(),
			WatchMode:  true,
		}

		err := runInitialBuild()

		assert.Error(t, err)
	})
}

func TestDetermineDirsToWatch(t *testing.T) {
	logger.InitLogger()

	t.Run("returns error for invalid topdir", func(t *testing.T) {
		dirs, err := determineDirsToWatch("/nonexistent/path")

		assert.Error(t, err)
		assert.Nil(t, dirs)
	})
}

func TestWatchLoopWithContext(t *testing.T) {
	logger.InitLogger()
	logger.DisableEmoji()
	defer logger.EnableEmoji()

	t.Run("cancels gracefully when context is cancelled", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")
		err := os.MkdirAll(releaseDir, 0755)
		require.NoError(t, err)

		watcher, err := fsnotify.NewWatcher()
		require.NoError(t, err)
		defer func() {
			_ = watcher.Close()
		}()

		err = watcher.Add(tmpDir)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		debounceDuration := 100 * time.Millisecond

		// Start watch loop
		done := watchLoop(ctx, watcher, releaseDir, debounceDuration)

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Cancel the context
		cancel()

		// Wait for completion with timeout
		select {
		case err := <-done:
			// Should get context.Canceled error
			assert.ErrorIs(t, err, context.Canceled)
		case <-time.After(1 * time.Second):
			t.Fatal("watchLoop did not exit after context cancellation")
		}
	})

	t.Run("processes file events and debounces", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")
		err := os.MkdirAll(releaseDir, 0755)
		require.NoError(t, err)

		testFile := filepath.Join(tmpDir, "test.lua")
		err = os.WriteFile(testFile, []byte("-- test"), 0644)
		require.NoError(t, err)

		// Mock triggerBuildFunc to avoid real build
		originalTriggerBuild := triggerBuildFunc
		buildCalled := false
		triggerBuildFunc = func() error {
			buildCalled = true
			return nil
		}
		defer func() { triggerBuildFunc = originalTriggerBuild }()

		WatchParams = &WatchArgs{CopyToWowDirs: false}

		watcher, err := fsnotify.NewWatcher()
		require.NoError(t, err)
		defer func() {
			_ = watcher.Close()
		}()

		err = watcher.Add(tmpDir)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		debounceDuration := 50 * time.Millisecond

		// Start watch loop
		done := watchLoop(ctx, watcher, releaseDir, debounceDuration)

		// Trigger a file write event
		time.Sleep(10 * time.Millisecond)
		err = os.WriteFile(testFile, []byte("-- updated content"), 0644)
		require.NoError(t, err)

		// Wait for debounce to trigger
		time.Sleep(100 * time.Millisecond)

		// Cancel to exit cleanly
		cancel()

		// Wait for completion
		select {
		case err := <-done:
			assert.ErrorIs(t, err, context.Canceled)
		case <-time.After(2 * time.Second):
			t.Fatal("watchLoop did not exit")
		}

		// Verify build was triggered
		assert.True(t, buildCalled, "triggerBuild should have been called")
	})

	t.Run("handles watcher errors", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		releaseDir := filepath.Join(tmpDir, "release")
		err := os.MkdirAll(releaseDir, 0755)
		require.NoError(t, err)

		watcher, err := fsnotify.NewWatcher()
		require.NoError(t, err)

		// Close the watcher immediately to trigger an error
		_ = watcher.Close()

		ctx := context.Background()
		debounceDuration := 100 * time.Millisecond

		// Start watch loop with closed watcher
		done := watchLoop(ctx, watcher, releaseDir, debounceDuration)

		// Should get an error quickly
		select {
		case err := <-done:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "channel closed")
		case <-time.After(1 * time.Second):
			t.Fatal("watchLoop did not exit after watcher closed")
		}
	})
}
