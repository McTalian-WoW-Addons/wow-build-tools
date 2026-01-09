package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

var tcl = logger.GetSubLog("TOC_CHECK")

func buildTocs(tocFilePaths []string) (tocFiles [](*Toc), err error) {
	for _, tocFilePath := range tocFilePaths {
		var tocFile *Toc
		tocFile, err = NewToc(tocFilePath)
		if err != nil {
			return
		}
		tocFiles = append(tocFiles, tocFile)
	}

	return
}

func checkName(addonDir string, tocFilePaths []string) (err error) {
	if !TocCheckParams.SkipNameCheck {
		addonName := DetermineProjectName(tocFilePaths)
		addonDirName := filepath.Base(addonDir)
		// TODO: Handle package-as pkgmeta field
		if addonName != addonDirName {
			err = fmt.Errorf("addon folder name '%s' does not match TOC file addon name '%s'", addonDirName, addonName)
		} else {
			tcl.Info("Completed name check: addon name is '%s', addon folder name is '%s'", addonName, addonDirName)
		}
	}
	return
}

func enumerateAddonLuaFiles(addonDir string) (addonLuaFiles []string, err error) {
	err = filepath.Walk(addonDir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if !info.IsDir() && filepath.Ext(path) == ".lua" {
			addonLuaFiles = append(addonLuaFiles, path)
		}
		return nil
	})

	return
}

func enumerateTocIncludedLuaFiles(addonDir string, tocFiles [](*Toc), ignoreFiles []string) (luaFilesInTocs []string, errs []error) {
	for _, tocFile := range tocFiles {
		includedFiles, e := getIncludedFiles(addonDir, tocFile, TocCheckParams.IgnoreFiles)
		if e != nil {
			errs = append(errs, fmt.Errorf("Error checking TOC file '%s': %v", tocFile.Filepath, e))
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
	return
}

func checkFiles(addonDir string, tocFiles [](*Toc)) (warns []string, errs []error) {
	if !TocCheckParams.SkipMissingFilesCheck {
		addonLuaFiles, err := enumerateAddonLuaFiles(addonDir)
		if err != nil {
			errs = append(errs, err)
			return
		}

		luaFilesInTocs, e := enumerateTocIncludedLuaFiles(addonDir, tocFiles, TocCheckParams.IgnoreFiles)
		if e != nil {
			errs = append(errs, e...)
			return
		}

		missingFromDisk, missingFromToc := difference(addonLuaFiles, luaFilesInTocs)
		for _, f := range missingFromDisk {
			errs = append(errs, fmt.Errorf("File '%s' is included in TOC but missing from disk", f))
		}
		for _, f := range missingFromToc {
			warns = append(warns, fmt.Sprintf("File '%s' is present on disk but missing from TOC", f))
		}

		tcl.Info("Completed missing files check:\n  %d files in TOC include tree but missing from addon folder\n  %d files in addon folder but missing from TOC include tree", len(missingFromDisk), len(missingFromToc))
	}

	return
}

func checkInterfaces(tocFiles [](*Toc)) (warns []string, err error) {
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
					warns = append(warns, fmt.Sprintf("TOC file '%s' uses interface version %d which is no longer a latest version", tocFile.Filepath, iface))
				}
			}

			// Check for any interfaces that are available but not in the TOC
			for iface := range iFaceMap {
				if !tocFaceMap[iface] {
					updateCount++
					warns = append(warns, fmt.Sprintf("TOC file '%s' is missing available interface version upgrade %d", tocFile.Filepath, iface))
				}
			}
		}

		tcl.Info("Completed interface version check: %d interface version updates available", updateCount)
	}
	return
}

func RunTocCheck() (err error) {
	defer func() {
		if err != nil {
			tcl.Error("TOC Check Error: %v", err)
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

	tocFiles, err := buildTocs(tocFilePaths)
	if err != nil {
		return
	}

	checkNameErr := checkName(addonDir, tocFilePaths)
	if checkNameErr != nil {
		checkErrors = append(checkErrors, checkNameErr.Error())
	}

	fileWarns, fileErrs := checkFiles(addonDir, tocFiles)
	if len(fileErrs) > 0 {
		for _, e := range fileErrs {
			checkErrors = append(checkErrors, e.Error())
		}
	}
	if len(fileWarns) > 0 {
		checkWarnings = append(checkWarnings, fileWarns...)
	}

	interfaceWarns, interfaceErr := checkInterfaces(tocFiles)
	if interfaceErr != nil {
		err = interfaceErr
		return
	}
	if len(interfaceWarns) > 0 {
		checkWarnings = append(checkWarnings, interfaceWarns...)
	}

	if len(checkErrors) > 0 {
		for _, e := range checkErrors {
			tcl.Error("%s", e)
		}
		err = fmt.Errorf("TOC check failed with %d errors", len(checkErrors))
		return
	}

	if len(checkWarnings) > 0 {
		for _, warning := range checkWarnings {
			tcl.Warn("%s", warning)
		}
	}

	tcl.Success("TOC check succeeded!")

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
		return nil, err
	}

	return tree.FlattenEntries(), nil
}
