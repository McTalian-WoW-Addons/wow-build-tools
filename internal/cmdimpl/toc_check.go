package cmdimpl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

type TocCheckArgs struct {
	IgnoreFiles           []string
	SkipInterfaceCheck    bool
	SkipMissingFilesCheck bool
	SkipNameCheck         bool
}

var TocCheckParams *TocCheckArgs = &TocCheckArgs{
	IgnoreFiles:           []string{},
	SkipInterfaceCheck:    false,
	SkipMissingFilesCheck: false,
	SkipNameCheck:         false,
}

var l = logger.DefaultLogger

func RunTocCheck() error {
	l := logger.GetSubLog("TOC_CHECK")
	if RootParams.LevelVerbose {
		l.SetLogLevel(logger.VERBOSE)
	} else if RootParams.LevelDebug {
		l.SetLogLevel(logger.DEBUG)
	} else {
		l.SetLogLevel(logger.INFO)
	}

	addonDir := TocParams.AddonDir
	checkErrors := []string{}
	checkWarnings := []string{}

	tocFilePaths, err := toc.FindTocFiles(addonDir)
	if err != nil {
		l.Error("TOC Error: %v", err)
		return err
	}

	var tocFiles [](*toc.Toc)
	for _, tocFilePath := range tocFilePaths {
		tocFile, err := toc.NewToc(tocFilePath)
		if err != nil {
			l.Error("Could not parse TOC file '%s': %v", tocFilePath, err)
			return err
		}
		tocFiles = append(tocFiles, tocFile)
	}

	addonName := toc.DetermineProjectName(tocFilePaths)
	addonDirName := filepath.Base(addonDir)
	if !TocCheckParams.SkipNameCheck {
		if addonName != addonDirName {
			checkErrors = append(checkErrors, fmt.Sprintf("Addon folder name '%s' does not match TOC file addon name '%s'", addonDirName, addonName))
		}

		l.Info("Completed name check: addon name is '%s', addon folder name is '%s'", addonName, addonDirName)
	}

	if !TocCheckParams.SkipMissingFilesCheck {
		var addonLuaFiles []string
		err = filepath.Walk(addonDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".lua" {
				addonLuaFiles = append(addonLuaFiles, path)
			}
			return nil
		})
		if err != nil {
			l.Error("Error walking addon directory '%s': %v", addonDir, err)
			return err
		}

		var luaFilesInTocs []string
		for _, tocFile := range tocFiles {
			includedFiles, err := getIncludedFiles(addonDirName, tocFile, TocCheckParams.IgnoreFiles)
			if err != nil {
				checkErrors = append(checkErrors, fmt.Sprintf("Error checking TOC file '%s': %v", tocFile.Filepath, err))
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
		flavorReleaseInfo := toc.FlavorReleaseInfo{
			IsBeta: TocParams.Beta,
			IsTest: TocParams.Ptr,
		}
		var updateCount int = 0
		for _, tocFile := range tocFiles {
			availableInterfaces, err := tocFile.CheckForInterfaceBumps(flavorReleaseInfo)
			if err != nil {
				l.Error("Error checking for interface bumps: %v", err)
				return err
			}

			var iFaceMap = make(map[int]bool)
			for _, iface := range availableInterfaces {
				iFaceMap[iface] = true
			}

			var tocFaceMap = make(map[int]bool)
			for _, iface := range tocFile.Interface {
				tocFaceMap[iface] = true
			}

			// Check for any interfaces in the TOC that are no longer the latest
			for _, iface := range tocFile.Interface {
				if !iFaceMap[iface] {
					updateCount++
					checkWarnings = append(checkWarnings, fmt.Sprintf("TOC file '%s' uses interface version %d which is no longer a latest version", tocFile.Filepath, iface))
				}
			}

			// Check for any interfaces that are available but not in the TOC
			for iface := range iFaceMap {
				if !tocFaceMap[iface] {
					updateCount++
					checkWarnings = append(checkWarnings, fmt.Sprintf("TOC file '%s' is missing available interface version upgrade %d", tocFile.Filepath, iface))
				}
			}
		}

		l.Info("Completed interface version check: %d interface version updates available", updateCount)
	}

	if len(checkErrors) > 0 {
		for _, err := range checkErrors {
			l.Error("%s", err)
		}
		return fmt.Errorf("TOC check failed with %d errors", len(checkErrors))
	}

	if len(checkWarnings) > 0 {
		for _, warning := range checkWarnings {
			l.Warn("%s", warning)
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

func getIncludedFiles(addonDir string, tocFile *toc.Toc, ignoreFiles []string) ([]string, error) {
	tree, err := tocFile.GetTocEntriesTree(addonDir, ignoreFiles, l)
	if err != nil {
		l.Error("Could not get TOC file tree for '%s': %v", tocFile.Filepath, err)
		return nil, err
	}

	return tree.FlattenEntries(), nil
}
