package build

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/McTalian/wow-build-tools/internal/flavor"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/spf13/viper"
)

type Flavor = flavor.Flavor

type LinkArgs struct {
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
			flavors = []Flavor{}
		}

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
			if forceLink {
				l.Debug("Removing existing symlink %s", filepath.Join(wowPath, "Interface", "AddOns", addonDir))
				err = os.RemoveAll(filepath.Join(wowPath, "Interface", "AddOns", addonDir))
				if err != nil && !os.IsNotExist(err) {
					l.Error("Error removing existing symlink: %v", err)
					return err
				}
			}
			source := filepath.Join(releaseDir, addonDir)
			target := filepath.Join(wowPath, "Interface", "AddOns", addonDir)
			l.Info("Linking %s to %s", source, target)
			err = os.Symlink(source, target)
			if err != nil {
				l.Error("Error creating symlink: %v", err)
				return err
			}
		}
	}
	return nil
}
