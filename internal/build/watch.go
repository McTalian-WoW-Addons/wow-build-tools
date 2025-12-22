package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/osutil"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type WatchArgs struct {
	CopyToWowDirs bool
}

var WatchParams = &WatchArgs{}

var addonDirs []string
var destinationPaths []string
var wowPaths map[string]string

// triggerBuildFunc is a function variable that can be replaced in tests for mocking.
var triggerBuildFunc = func(done chan error) {
	triggerBuild(done)
}

// runInitialBuildFunc is a function variable that can be replaced in tests for mocking.
var runInitialBuildFunc = func() error {
	return runInitialBuild()
}

// setupWatchEnvironmentFunc is a function variable that can be replaced in tests for mocking.
var setupWatchEnvironmentFunc = func(releaseDir string, copyToWowDirs bool) error {
	return setupWatchEnvironment(releaseDir, copyToWowDirs)
}

// determineDirsToWatchFunc is a function variable that can be replaced in tests for mocking.
var determineDirsToWatchFunc = func(topdir string) ([]string, error) {
	return determineDirsToWatch(topdir)
}

// setupWatcherFunc is a function variable that can be replaced in tests for mocking.
var setupWatcherFunc = func(dirsToWatch []string) (*fsnotify.Watcher, error) {
	return setupWatcher(dirsToWatch)
}

func copyToWow(l *logger.Logger, done chan error) {
	if WatchParams.CopyToWowDirs {
		l.Info("Copying to WoW directories...")
		lg := logger.NewLogGroup("Copy to WoW Directories", l)

		var copyWg sync.WaitGroup
		for _, path := range destinationPaths {
			copyWg.Add(1)
			go func(path string) {
				defer copyWg.Done()
				interfaceDir := filepath.Join(path, "Interface", "AddOns")
				if _, err := os.Stat(interfaceDir); os.IsNotExist(err) {
					err = os.MkdirAll(interfaceDir, os.ModePerm)
					if err != nil {
						l.Error("Error creating directory %s: %v", interfaceDir, err)
						done <- err
						return
					}
				}

				for _, dir := range addonDirs {
					src := filepath.Join(BuildParams.ReleaseDir, dir)
					dst := filepath.Join(interfaceDir, dir)
					l.Debug("Copying %s to %s", src, dst)
					err := copyDir(src, dst)
					if err != nil {
						l.Error("Error copying %s to %s: %v", src, dst, err)
						done <- err
					}
				}
			}(path)
		}

		copyWg.Wait()
		lg.Flush()
	}
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Skip copy if destination exists and modification times are equal.
	if dfi, err := os.Stat(dst); err == nil {
		if sfi.ModTime().Equal(dfi.ModTime()) && sfi.Size() == dfi.Size() {
			return nil
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	// Ensure the destination directory exists.
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	// Copy file contents.
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Preserve modification time.
	return os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
}

// copyDir recursively copies a directory from src to dst concurrently.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(entries))

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			wg.Add(1)
			// Recurse into directories concurrently.
			go func(srcDir, dstDir string) {
				defer wg.Done()
				if err := copyDir(srcDir, dstDir); err != nil {
					errCh <- err
				}
			}(srcPath, dstPath)
		} else {
			wg.Add(1)
			// Copy files concurrently.
			go func(srcFile, dstFile string) {
				defer wg.Done()
				if err := copyFile(srcFile, dstFile); err != nil {
					errCh <- err
				}
			}(srcPath, dstPath)
		}
	}

	wg.Wait()
	close(errCh)

	// Return the first error encountered, if any.
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func triggerBuild(done chan error) {
	buildArgs := &BuildArgs{
		TopDir:         BuildParams.TopDir,
		ReleaseDir:     BuildParams.ReleaseDir,
		SkipChangelog:  true,
		SkipUpload:     true,
		SkipZip:        true,
		KeepPackageDir: true,
		WatchMode:      true,
	}
	logger.Clear()
	err := Build(buildArgs)
	if err != nil {
		logger.Error("Error running build command: %v", err)
		done <- err
		return
	}
	fmt.Println()
}

func determineDirsToWatch(topdir string) (dirsToWatch []string, err error) {
	defer func() {
		if err != nil {
			l.Error("Error determining directories to watch: %v", err)
		}
	}()

	var tree []string
	tree, err = toc.GetTocFileTree(topdir)
	if err != nil {
		return
	}

	l.Verbose("Tree: %v", tree)

	var entries []string
	for _, file := range tree {
		if filepath.Ext(file) == ".xml" {
			l.Verbose("Walking XML file: %s", file)
			var xmlEntries []string
			xmlEntries, err = toc.WalkXmlFile(file, l)
			if err != nil {
				return
			}
			entries = append(entries, xmlEntries...)
		} else {
			l.Verbose("Adding file: %s", file)
			entries = append(entries, file)
		}
	}

	var dirsToWatchSet = make(map[string]bool)
	for _, entry := range entries {
		if f, err := os.Stat(entry); err == nil {
			if f.IsDir() {
				dirsToWatchSet[entry] = true
			} else {
				dirsToWatchSet[filepath.Dir(entry)] = true
			}
		}
	}

	for dir := range dirsToWatchSet {
		dirsToWatch = append(dirsToWatch, dir)
	}

	return
}

func onDebounceExpired(done chan error, releaseDir string) {
	l.Debug("Debounced change detected, triggering build...")
	triggerBuildFunc(done)

	if WatchParams.CopyToWowDirs {
		l.Info("Build complete, determining outputs to copy...")
		dirEntries, err := os.ReadDir(releaseDir)
		if err != nil {
			l.Error("Error reading release directory: %v", err)
			done <- err
		}

		addonDirs = []string{}
		for _, entry := range dirEntries {
			if entry.IsDir() {
				addonDirs = append(addonDirs, entry.Name())
			}
		}

		copyToWow(l, done)
	}
}

// setupWatchEnvironment prepares the environment for watching, including creating directories
// and handling WSL-specific warnings.
func setupWatchEnvironment(releaseDir string, copyToWowDirs bool) error {
	if _, err := os.Stat(releaseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(releaseDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create release directory: %w", err)
		}
	}

	if osutil.IsWSL() && !copyToWowDirs {
		winPath, err := osutil.GetWindowsPath(releaseDir)
		if err != nil {
			return fmt.Errorf("failed to get Windows path: %w", err)
		}

		l.Warn("To create symlinks to your release directory in WSL, run this command in Windows in an elevated command prompt:")
		l.Warn("wow-build-tools.exe link -w \"%s\"", winPath)
	}

	// Clean the release directory
	if err := os.RemoveAll(releaseDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean release directory: %w", err)
	}

	return nil
}

// runInitialBuild performs the initial build before starting the watch loop.
func runInitialBuild() error {
	initialBuildChan := make(chan error, 1)
	triggerBuild(initialBuildChan)
	close(initialBuildChan)

	for e := range initialBuildChan {
		if e != nil {
			return e
		}
	}
	return nil
}

// loadWowPaths loads and validates WoW installation paths from configuration.
func loadWowPaths() ([]string, error) {
	wowPaths = viper.GetStringMapString("wowPath")
	if len(wowPaths) <= 1 {
		return nil, fmt.Errorf("no WoW paths configured, please run 'wow-build-tools config' to set up at least one WoW installation path")
	}

	paths := make([]string, 0, len(wowPaths)-1)
	for key, path := range wowPaths {
		if key == "base" {
			continue
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// setupWatcher creates and configures the file system watcher for the given directories.
func setupWatcher(dirsToWatch []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	for _, dir := range dirsToWatch {
		if err := watcher.Add(dir); err != nil {
			_ = watcher.Close()
			return nil, fmt.Errorf("failed to add directory to watcher: %w", err)
		}
		l.Debug("Watching directory: %s", dir)
	}

	return watcher, nil
}

// shouldProcessEvent determines if a file system event should trigger a rebuild.
func shouldProcessEvent(event fsnotify.Event, releaseDir string) bool {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}

	if strings.Contains(event.Name, releaseDir) {
		l.Debug("Skipping change event on release directory")
		return false
	}

	return true
}

// watchLoop runs the main file watching loop, handling events and debouncing.
// It accepts a context for cancellation, making it testable and allowing graceful shutdown.
func watchLoop(ctx context.Context, watcher *fsnotify.Watcher, releaseDir string, debounceDuration time.Duration) chan error {
	done := make(chan error, 1) // Buffered to prevent goroutine leak

	go func() {
		defer close(done)

		var debounceTimer *time.Timer

		// Initialize the debounce timer (stopped)
		debounceTimer = time.AfterFunc(debounceDuration, func() {
			onDebounceExpired(done, releaseDir)
			l.Info("Watching for changes... Press Ctrl+C to stop.")
		})
		debounceTimer.Stop()

		for {
			select {
			case <-ctx.Done():
				// Context cancelled - clean shutdown
				debounceTimer.Stop()
				done <- ctx.Err()
				return

			case event, ok := <-watcher.Events:
				if !ok {
					done <- fmt.Errorf("watcher events channel closed")
					return
				}

				if shouldProcessEvent(event, releaseDir) {
					debounceTimer.Reset(debounceDuration)
					l.Debug("Change %s detected on %s, debouncing...", event.Op, event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					done <- fmt.Errorf("watcher errors channel closed")
					return
				}
				l.Error("Watcher error: %v", err)
				done <- err
				return
			}
		}
	}()

	return done
}

// WatchBuild watches for file changes and rebuilds the addon automatically.
// It accepts a context for cancellation, allowing graceful shutdown via Ctrl+C or testing.
func WatchBuild(ctx context.Context) (err error) {
	l := logger.GetSubLog("WATCH")

	defer func() {
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			l.Error("Watch build failed: %v", err)
		}
	}()

	topdir := BuildParams.TopDir
	releaseDir := BuildParams.ReleaseDir

	// Setup environment
	if err = setupWatchEnvironmentFunc(releaseDir, WatchParams.CopyToWowDirs); err != nil {
		return
	}

	// Run initial build
	if err = runInitialBuildFunc(); err != nil {
		return
	}

	// Load WoW paths if needed
	if WatchParams.CopyToWowDirs {
		destinationPaths, err = loadWowPaths()
		if err != nil {
			return err
		}
	}

	// Determine directories to watch
	dirsToWatch, err := determineDirsToWatchFunc(topdir)
	if err != nil {
		return
	}

	// Setup file system watcher
	watcher, err := setupWatcherFunc(dirsToWatch)
	if err != nil {
		return
	}
	defer func() { _ = watcher.Close() }()

	// Start watch loop with context
	debounceDuration := 500 * time.Millisecond
	done := watchLoop(ctx, watcher, releaseDir, debounceDuration)

	l.Info("Watching for changes... Press Ctrl+C to stop.")

	// Wait for completion, error, or cancellation
	select {
	case err = <-done:
		return err
	case <-ctx.Done():
		// Context cancelled from outside (e.g., Ctrl+C)
		<-done // Wait for watchLoop to finish cleanup
		return ctx.Err()
	}
}
