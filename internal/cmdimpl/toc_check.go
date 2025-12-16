package cmdimpl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

type TocCheckArgs struct {
	AddonDir    string
	IgnoreFiles []string

	SkipInterfaceCheck    bool
	SkipMissingFilesCheck bool
	SkipNameCheck         bool

	LevelVerbose bool
	LevelDebug   bool
}

var addonDir string

var TocCheckParams *TocCheckArgs = &TocCheckArgs{
	AddonDir:    ".",
	IgnoreFiles: []string{},

	SkipInterfaceCheck:    false,
	SkipMissingFilesCheck: false,
	SkipNameCheck:         false,

	LevelVerbose: false,
	LevelDebug:   false,
}

var l = logger.DefaultLogger

func RunTocCheck() error {
	checkErrors := []string{}
	checkWarnings := []string{}

	tocFilePaths, err := toc.FindTocFiles(TocCheckParams.AddonDir)
	if err != nil {
		l.Error("TOC Error: %v", err)
		return err
	}

	addonName := toc.DetermineProjectName(tocFilePaths)
	addonDirName := filepath.Base(TocCheckParams.AddonDir)
	if !TocCheckParams.SkipNameCheck {
		if addonName != addonDirName {
			checkErrors = append(checkErrors, fmt.Sprintf("Addon folder name '%s' does not match TOC file addon name '%s'", addonDirName, addonName))
		}

		l.Info("Completed name check: addon name is '%s', addon folder name is '%s'", addonName, addonDirName)
	}

	if !TocCheckParams.SkipMissingFilesCheck {
		var addonLuaFiles []string
		err = filepath.Walk(TocCheckParams.AddonDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".lua" {
				addonLuaFiles = append(addonLuaFiles, path)
			}
			return nil
		})
		if err != nil {
			l.Error("Error walking addon directory '%s': %v", TocCheckParams.AddonDir, err)
			return err
		}

		var luaFilesInTocs []string
		for _, tocFilePath := range tocFilePaths {
			includedFiles, err := getIncludedFiles(addonDirName, tocFilePath, TocCheckParams.IgnoreFiles)
			if err != nil {
				checkErrors = append(checkErrors, fmt.Sprintf("Error checking TOC file '%s': %v", tocFilePath, err))
				continue
			}

			for _, f := range includedFiles {
				if filepath.Ext(f) == ".lua" {
					luaFilesInTocs = append(luaFilesInTocs, f)
				}
			}
		}

		missingFromDisk, missingFromToc := difference(addonLuaFiles, luaFilesInTocs)
		for _, f := range missingFromDisk {
			checkErrors = append(checkErrors, fmt.Sprintf("File '%s' is included in TOC but missing from disk", f))
		}
		for _, f := range missingFromToc {
			checkWarnings = append(checkWarnings, fmt.Sprintf("File '%s' is present on disk but missing from TOC", f))
		}

		l.Info("Completed missing files check:\n  %d files in TOC include tree but missing from addon folder\n  %d files in addon folder but missing from TOC include tree", len(missingFromDisk), len(missingFromToc))
	}

	if !TocCheckParams.SkipInterfaceCheck {
		// TODO
	}

	if len(checkErrors) > 0 {
		for _, err := range checkErrors {
			l.Error(err)
		}
		return fmt.Errorf("TOC check failed with %d errors", len(checkErrors))
	}

	if len(checkWarnings) > 0 {
		for _, warning := range checkWarnings {
			l.Warn(warning)
		}
	}

	l.Success("TOC check succeeded!")

	return nil
}

func difference(a, b []string) (missingA []string, missingB []string) {
	missingA = []string{}
	missingB = []string{}
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	ma := make(map[string]struct{}, len(a))
	for _, x := range a {
		ma[x] = struct{}{}
	}
	for _, x := range a {
		if _, found := mb[x]; !found {
			missingB = append(missingB, x)
		}
	}
	for _, x := range b {
		if _, found := ma[x]; !found {
			missingA = append(missingA, x)
		}
	}
	return
}

func getIncludedFiles(addonDir, tocFilePath string, ignoreFiles []string) ([]string, error) {
	tocInstance, err := toc.NewToc(tocFilePath)
	if err != nil {
		l.Error("Could not parse TOC file '%s': %v", tocFilePath, err)
		return nil, err
	}

	tree, err := tocInstance.GetTocEntriesTree(addonDir, ignoreFiles, l)
	if err != nil {
		l.Error("Could not get TOC file tree for '%s': %v", tocFilePath, err)
		return nil, err
	}

	return tree.FlattenEntries(), nil
}
