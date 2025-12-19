package build

import (
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

func WatchBuild() (err error) {
	l := logger.GetSubLog("WATCH")

	defer func() {
		if err != nil {
			l.Error("Watch build failed: %v", err)
		}
	}()

	topdir := BuildParams.TopDir
	releaseDir := BuildParams.ReleaseDir

	if _, e := os.Stat(releaseDir); os.IsNotExist(e) {
		err = os.MkdirAll(releaseDir, os.ModePerm)
		if err != nil {
			return
		}
	}

	if osutil.IsWSL() && !WatchParams.CopyToWowDirs {
		var winPath string
		winPath, err = osutil.GetWindowsPath(releaseDir)
		if err != nil {
			return
		}

		l.Warn("To create symlinks to your release directory in WSL, run this command in Windows in an elevated command prompt:")
		l.Warn("wow-build-tools.exe link -w \"%s\"", winPath)
	}

	err = os.RemoveAll(BuildParams.ReleaseDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialBuildChan := make(chan error, 1)
	triggerBuild(initialBuildChan)
	close(initialBuildChan)
	for e := range initialBuildChan {
		if e != nil {
			err = e
			return
		}
	}

	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer func() { _ = watcher.Close() }()

	if WatchParams.CopyToWowDirs {
		wowPaths = viper.GetStringMapString("wowPath")
		if len(wowPaths) <= 1 {
			err = fmt.Errorf("no WoW paths configured, please run 'wow-build-tools config' to set up at least one WoW installation path")
			return
		}
		destinationPaths = make([]string, 0, len(wowPaths)-1)
		for key, path := range wowPaths {
			if key == "base" {
				continue
			}
			destinationPaths = append(destinationPaths, path)
		}
	}

	debounceDuration := 500 * time.Millisecond
	var debounceTimer *time.Timer

	done := make(chan error)
	go func() {
		// Reset the debounce timer. When it fires, run the build.
		debounceTimer = time.AfterFunc(debounceDuration, func() {
			// It's a good idea to ensure builds don’t run concurrently.
			// You can use a mutex, a channel, or a boolean flag as in your current implementation.
			l.Debug("Debounced change detected, triggering build...")
			triggerBuild(done)

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

			l.Info("Watching for changes... Press Ctrl+C to stop.")
		})
		debounceTimer.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- fmt.Errorf("error reading from watcher")
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					if strings.Contains(event.Name, releaseDir) {
						l.Debug("Skipping change event on release directory")
						continue
					}

					debounceTimer.Reset(debounceDuration)
					l.Debug("Change %s detected on %s, debouncing...", event.Op, event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					done <- fmt.Errorf("error reading from watcher")
				}
				l.Error("Watcher error: %v", err)
				done <- err
			}
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
		err = watcher.Add(dir)
		if err != nil {
			return
		}
		l.Debug("Watching directory: %s", dir)
	}

	l.Info("Watching for changes... Press Ctrl+C to stop.")

	<-done

	return nil
}
