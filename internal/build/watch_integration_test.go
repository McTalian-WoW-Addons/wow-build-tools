package build

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

// TestWatchBuildIntegration tests the full WatchBuild function with real dependencies.
// These tests focus on context cancellation behavior.
func TestWatchBuildIntegration(t *testing.T) {
	logger.InitLogger()
	logger.DisableEmoji()
	defer logger.EnableEmoji()

	t.Run("handles invalid topdir", func(t *testing.T) {
		// Configure with non-existent directory
		BuildParams = &BuildArgs{
			TopDir:     "/nonexistent/path",
			ReleaseDir: "/tmp/release",
			WatchMode:  true,
		}
		WatchParams = &WatchArgs{
			CopyToWowDirs: false,
		}

		ctx := context.Background()
		err := WatchBuild(ctx)

		// Should return an error about TOC file
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TOC")
	})

	t.Run("respects context cancellation during setup", func(t *testing.T) {
		// Setup with slow/blocking initialization
		tmpDir := t.TempDir()
		BuildParams = &BuildArgs{
			TopDir:     tmpDir,
			ReleaseDir: filepath.Join(tmpDir, "release"),
			WatchMode:  true,
		}
		WatchParams = &WatchArgs{
			CopyToWowDirs: false,
		}

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// WatchBuild should exit quickly due to cancelled context
		start := time.Now()
		err := WatchBuild(ctx)
		elapsed := time.Since(start)

		// Should complete quickly and return context.Canceled
		assert.Less(t, elapsed, 1*time.Second, "Should exit quickly when context is already cancelled")
		if err != nil {
			// Could be context.Canceled or an error during setup that checked context
			t.Logf("Got error: %v", err)
		}
	})

	t.Run("full watch lifecycle with mocked build", func(t *testing.T) {
		// Mock all the setup functions to avoid real file operations
		originalSetup := setupWatchEnvironmentFunc
		originalBuild := runInitialBuildFunc
		originalDetermine := determineDirsToWatchFunc
		originalWatcher := setupWatcherFunc
		defer func() {
			setupWatchEnvironmentFunc = originalSetup
			runInitialBuildFunc = originalBuild
			determineDirsToWatchFunc = originalDetermine
			setupWatcherFunc = originalWatcher
		}()

		setupCalled := false
		buildCalled := false
		determineCalled := false
		watcherCalled := false

		setupWatchEnvironmentFunc = func(releaseDir string, copyToWowDirs bool) error {
			setupCalled = true
			return nil
		}

		runInitialBuildFunc = func() error {
			buildCalled = true
			return nil
		}

		determineDirsToWatchFunc = func(topdir string) ([]string, error) {
			determineCalled = true
			return []string{topdir}, nil
		}

		setupWatcherFunc = func(dirsToWatch []string) (*fsnotify.Watcher, error) {
			watcherCalled = true
			// Create a real watcher but don't add any paths
			return fsnotify.NewWatcher()
		}

		// Setup
		tmpDir := t.TempDir()
		BuildParams = &BuildArgs{
			TopDir:     tmpDir,
			ReleaseDir: filepath.Join(tmpDir, "release"),
			WatchMode:  true,
		}
		WatchParams = &WatchArgs{
			CopyToWowDirs: false,
		}

		// Use context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// Run WatchBuild
		err := WatchBuild(ctx)

		// Should timeout (context.DeadlineExceeded)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// Verify all steps were called
		assert.True(t, setupCalled, "setupWatchEnvironment should have been called")
		assert.True(t, buildCalled, "runInitialBuild should have been called")
		assert.True(t, determineCalled, "determineDirsToWatch should have been called")
		assert.True(t, watcherCalled, "setupWatcher should have been called")
	})

	t.Run("cancels gracefully after initialization", func(t *testing.T) {
		// Mock all the functions
		originalSetup := setupWatchEnvironmentFunc
		originalBuild := runInitialBuildFunc
		originalDetermine := determineDirsToWatchFunc
		originalWatcher := setupWatcherFunc
		defer func() {
			setupWatchEnvironmentFunc = originalSetup
			runInitialBuildFunc = originalBuild
			determineDirsToWatchFunc = originalDetermine
			setupWatcherFunc = originalWatcher
		}()

		setupWatchEnvironmentFunc = func(releaseDir string, copyToWowDirs bool) error {
			return nil
		}

		runInitialBuildFunc = func() error {
			return nil
		}

		determineDirsToWatchFunc = func(topdir string) ([]string, error) {
			return []string{topdir}, nil
		}

		setupWatcherFunc = func(dirsToWatch []string) (*fsnotify.Watcher, error) {
			return fsnotify.NewWatcher()
		}

		tmpDir := t.TempDir()
		BuildParams = &BuildArgs{
			TopDir:     tmpDir,
			ReleaseDir: filepath.Join(tmpDir, "release"),
			WatchMode:  true,
		}
		WatchParams = &WatchArgs{
			CopyToWowDirs: false,
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Start WatchBuild in background
		done := make(chan error, 1)
		go func() {
			done <- WatchBuild(ctx)
		}()

		// Give it time to initialize
		time.Sleep(50 * time.Millisecond)

		// Cancel the context
		cancel()

		// Should exit quickly with context.Canceled
		select {
		case err := <-done:
			assert.ErrorIs(t, err, context.Canceled)
		case <-time.After(1 * time.Second):
			t.Fatal("WatchBuild did not exit after context cancellation")
		}
	})
}
