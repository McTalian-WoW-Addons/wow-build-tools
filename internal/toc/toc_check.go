package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

func RunTocCheck() (err error) {
	l := logger.GetSubLog("TOC_CHECK")

	defer func() {
		if err != nil {
			l.Error("TOC Check Error: %v", err)
		}
	}()

	addonDir := TocParams.AddonDir
	checkErrors := []string{}
	checkWarnings := []string{}

	var tocFilePaths []string
	tocFilePaths, err = FindTocFiles(addonDir)
	if err != nil {
		return
	}

	var tocFiles [](*Toc)
	for _, tocFilePath := range tocFilePaths {
		var tocFile *Toc
		tocFile, err = NewToc(tocFilePath)
		if err != nil {
			return
		}
		tocFiles = append(tocFiles, tocFile)
	}

	addonName := DetermineProjectName(tocFilePaths)
	addonDirName := filepath.Base(addonDir)
	if !TocCheckParams.SkipNameCheck {
		if addonName != addonDirName {
			checkErrors = append(checkErrors, fmt.Sprintf("Addon folder name '%s' does not match TOC file addon name '%s'", addonDirName, addonName))
		}

		l.Info("Completed name check: addon name is '%s', addon folder name is '%s'", addonName, addonDirName)
	}

	if !TocCheckParams.SkipMissingFilesCheck {
		var addonLuaFiles []string

		err = filepath.Walk(addonDir, func(path string, info os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			if !info.IsDir() && filepath.Ext(path) == ".lua" {
				addonLuaFiles = append(addonLuaFiles, path)
			}
			return nil
		})
		if err != nil {
			return
		}

		var luaFilesInTocs []string
		for _, tocFile := range tocFiles {
			includedFiles, e := getIncludedFiles(addonDir, tocFile, TocCheckParams.IgnoreFiles)
			if e != nil {
				checkErrors = append(checkErrors, fmt.Sprintf("Error checking TOC file '%s': %v", tocFile.Filepath, e))
				continue
			}

			for _, f := range includedFiles {
				if filepath.Ext(f) == ".lua" {
					// detect windows path separators and convert to unix style for comparison
					if os.PathSeparator == '\\' {
						f = filepath.ToSlash(f)
					} else {
						f = strings.ReplaceAll(f, "\\", "/")
					}
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
		flavorReleaseInfo := FlavorReleaseInfo{
			IsBeta: TocParams.Beta,
			IsTest: TocParams.Ptr,
		}
		var updateCount = 0
		for _, tocFile := range tocFiles {
			var availableInterfaces map[Product]int
			availableInterfaces, err = tocFile.CheckForInterfaceBumps(flavorReleaseInfo)
			if err != nil {
				return
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
		for _, e := range checkErrors {
			l.Error("%s", e)
		}
		err = fmt.Errorf("TOC check failed with %d errors", len(checkErrors))
		return
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

func getIncludedFiles(addonDir string, tocFile *Toc, ignoreFiles []string) ([]string, error) {
	tree, err := tocFile.GetTocEntriesTree(addonDir, ignoreFiles, l)
	if err != nil {
		l.Error("Could not get TOC file tree for '%s': %v", tocFile.Filepath, err)
		return nil, err
	}

	return tree.FlattenEntries(), nil
}
