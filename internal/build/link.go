package build

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/McTalian/wow-build-tools/internal/flavor"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/spf13/viper"
)

type Flavor = flavor.Flavor

type LinkArgs struct {
	AllFlavors  bool
	Force       bool
	OnlyFlavors []string

	WSLPathToAddonReleaseDir string
}

var LinkParams = &LinkArgs{}

func Link() error {
	l := logger.GetSubLog("link")

	onlyFlavors := LinkParams.OnlyFlavors
	forceLink := LinkParams.Force
	wslPathToAddonReleaseDir := LinkParams.WSLPathToAddonReleaseDir
	releaseDir := BuildParams.ReleaseDir

	flavors := make([]Flavor, len(flavor.KnownFlavors))
	copy(flavors, flavor.KnownFlavors)
	// Filter flavors if specific ones are requested
	if len(onlyFlavors) > 0 {
		filteredFlavors := make(map[Flavor]bool, len(flavor.KnownFlavors))
		for _, flavorStr := range onlyFlavors {
			f := flavor.FromId(flavorStr)
			if !slices.Contains(flavor.KnownFlavors, f) {
				l.Warn("Unknown flavor: %s", flavorStr)
				continue
			}
			filteredFlavors[f] = true
		}

		// If specific flavors were requested, we'll rebuild the
		// flavors slice to only include those
		flavors = []Flavor{}
		for filteredFlavor, include := range filteredFlavors {
			if include {
				flavors = append(flavors, filteredFlavor)
			}
		}
	}

	wowPaths := viper.GetStringMapString("wowPath")
	if len(wowPaths) <= 1 {
		l.Error("Please run `wow-build-tools config` to set up your World of Warcraft paths.")
		return fmt.Errorf("no World of Warcraft paths set")
	}

	if wslPathToAddonReleaseDir != "" {
		l.Debug("Using wslPathToAddonReleaseDir to determine WSL path to addon release directory")
		releaseDir = wslPathToAddonReleaseDir
	}

	l.Debug("Creating symlinks pointing to addons in %s", releaseDir)
	dirEntries, err := os.ReadDir(releaseDir)
	if err != nil {
		l.Error("Error reading release directory: %v", err)
		return err
	}

	addonDirs := []string{}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			addonDirs = append(addonDirs, entry.Name())
		}
	}

	if len(addonDirs) == 0 {
		l.Error("No addon directories found in release directory, please run a build first")
		return fmt.Errorf("no addon directories found in release directory")
	}

	if !LinkParams.AllFlavors {
		compatibleFlavors, err := getCompatibleInstallFlavors(BuildParams.TopDir, releaseDir, addonDirs)
		if err != nil {
			l.Warn("Unable to derive compatible WoW clients from TOC interface versions: %v", err)
			l.Warn("Falling back to linking for all selected client installations")
		} else if len(compatibleFlavors) > 0 {
			flavors = intersectFlavors(flavors, compatibleFlavors)
			l.Debug("TOC-compatible WoW clients: %v", flavorIds(compatibleFlavors))
		}
	}

	linksCreated := 0

	for k, wowPath := range wowPaths {
		if k == "base" {
			continue
		}

		if !slices.Contains(flavors, flavor.FromId(k)) {
			l.Debug("Skipping flavor %s", k)
			continue
		}

		if _, err := os.Stat(filepath.Join(wowPath)); os.IsNotExist(err) {
			l.Error("World of Warcraft path %s does not exist", wowPath)
			return err
		}

		if _, err := os.Stat(filepath.Join(wowPath, "Interface", "AddOns")); os.IsNotExist(err) {
			l.Warn("No AddOns directory found in %s, creating it", wowPath)
			err = os.MkdirAll(filepath.Join(wowPath, "Interface", "AddOns"), 0755)
			if err != nil {
				l.Error("Error creating AddOns directory: %v", err)
				return err
			}
		}
		for _, addonDir := range addonDirs {
			source := filepath.Join(releaseDir, addonDir)
			target := filepath.Join(wowPath, "Interface", "AddOns", addonDir)

			if linkErr := handleExistingTarget(target, forceLink, l); linkErr != nil {
				return linkErr
			}

			l.Info("Linking %s to %s", source, target)
			err = os.Symlink(source, target)
			if err != nil {
				l.Error("Error creating symlink: %v", err)
				return err
			}
			linksCreated++
		}
	}

	if linksCreated == 0 {
		l.Warn("No compatible World of Warcraft client installations were selected for linking")
	}

	return nil
}

func getCompatibleInstallFlavors(topDir, releaseDir string, addonDirs []string) ([]Flavor, error) {
	compatibleFlavors, topDirErr := getCompatibleInstallFlavorsFromDir(topDir)
	if topDirErr == nil && len(compatibleFlavors) > 0 {
		return compatibleFlavors, nil
	}

	compatibleFlavors, releaseDirErr := getCompatibleInstallFlavorsFromReleaseDir(releaseDir, addonDirs)
	if releaseDirErr == nil && len(compatibleFlavors) > 0 {
		return compatibleFlavors, nil
	}

	if topDirErr != nil && releaseDirErr != nil {
		return nil, fmt.Errorf("topDir lookup failed (%s): %v; releaseDir lookup failed (%s): %v", topDir, topDirErr, releaseDir, releaseDirErr)
	}

	return nil, fmt.Errorf("no compatible install flavors were derived from TOCs in %s or %s", topDir, releaseDir)
}

func getCompatibleInstallFlavorsFromDir(dir string) ([]Flavor, error) {
	_, tocFiles, err := getTocFiles(dir)
	if err != nil {
		return nil, err
	}

	return collectCompatibleFlavors(tocFiles), nil
}

func getCompatibleInstallFlavorsFromReleaseDir(releaseDir string, addonDirs []string) ([]Flavor, error) {
	var allTocs []*toc.Toc

	_, releaseRootTocs, err := getTocFiles(releaseDir)
	if err == nil {
		allTocs = append(allTocs, releaseRootTocs...)
	}

	for _, addonDir := range addonDirs {
		addonPath := filepath.Join(releaseDir, addonDir)
		_, addonTocs, addonErr := getTocFiles(addonPath)
		if addonErr != nil {
			continue
		}
		allTocs = append(allTocs, addonTocs...)
	}

	if len(allTocs) == 0 {
		return nil, fmt.Errorf("no TOC files found in release directory %s", releaseDir)
	}

	return collectCompatibleFlavors(allTocs), nil
}

func collectCompatibleFlavors(tocFiles []*toc.Toc) []Flavor {

	compatibleFlavors := []Flavor{}
	seenFlavorIds := make(map[string]bool)
	for _, tocFile := range tocFiles {
		for _, installFlavor := range tocFile.CompatibleInstallFlavors() {
			if seenFlavorIds[installFlavor.Id] {
				continue
			}

			compatibleFlavors = append(compatibleFlavors, installFlavor)
			seenFlavorIds[installFlavor.Id] = true
		}
	}

	return compatibleFlavors
}

func intersectFlavors(left []Flavor, right []Flavor) []Flavor {
	rightById := make(map[string]bool, len(right))
	for _, rightFlavor := range right {
		rightById[rightFlavor.Id] = true
	}

	intersection := []Flavor{}
	for _, leftFlavor := range left {
		if rightById[leftFlavor.Id] {
			intersection = append(intersection, leftFlavor)
		}
	}

	return intersection
}

func flavorIds(flavors []Flavor) []string {
	ids := make([]string, 0, len(flavors))
	for _, compatibleFlavor := range flavors {
		ids = append(ids, compatibleFlavor.Id)
	}

	return ids
}

func handleExistingTarget(target string, forceLink bool, l *logger.Logger) error {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		l.Error("Error checking existing target %s: %v", target, err)
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		l.Debug("Replacing existing symlink %s", target)
		if removeErr := os.Remove(target); removeErr != nil {
			l.Error("Error removing existing symlink %s: %v", target, removeErr)
			return removeErr
		}
		return nil
	}

	if !forceLink {
		err = fmt.Errorf("destination exists and is not a symlink: %s (use --force to overwrite)", target)
		l.Error("%v", err)
		return err
	}

	l.Debug("--force enabled, removing existing non-symlink target %s", target)
	if removeErr := os.RemoveAll(target); removeErr != nil {
		l.Error("Error removing existing target %s: %v", target, removeErr)
		return removeErr
	}

	return nil
}
