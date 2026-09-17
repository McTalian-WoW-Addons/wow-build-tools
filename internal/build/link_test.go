package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetLinkParams() {
	LinkParams = &LinkArgs{}
}

func TestLinkUsesTocCompatibleClients(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"

	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	classicInstall := filepath.Join(t.TempDir(), "classic")
	classicEraInstall := filepath.Join(t.TempDir(), "classic-era")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	require.NoError(t, os.MkdirAll(classicInstall, 0755))
	require.NoError(t, os.MkdirAll(classicEraInstall, 0755))

	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)
	viper.Set("wowPath.classic", classicInstall)
	viper.Set("wowPath.classicEra", classicEraInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(retailInstall, "Interface", "AddOns", addonName))
	assert.NoFileExists(t, filepath.Join(classicInstall, "Interface", "AddOns", addonName))
	assert.NoFileExists(t, filepath.Join(classicEraInstall, "Interface", "AddOns", addonName))
}

func TestLinkAllFlavorsBypassesTocFiltering(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"

	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	classicInstall := filepath.Join(t.TempDir(), "classic")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	require.NoError(t, os.MkdirAll(classicInstall, 0755))

	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)
	viper.Set("wowPath.classic", classicInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir
	LinkParams.AllFlavors = true

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(retailInstall, "Interface", "AddOns", addonName))
	assertSymlinkExists(t, filepath.Join(classicInstall, "Interface", "AddOns", addonName))
}

func TestLinkUsesWslReleaseDirForTocCompatibility(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := t.TempDir()
	addonName := "TestAddon"
	addonReleaseDir := filepath.Join(releaseDir, addonName)

	require.NoError(t, os.MkdirAll(addonReleaseDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(addonReleaseDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	classicInstall := filepath.Join(t.TempDir(), "classic")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	require.NoError(t, os.MkdirAll(classicInstall, 0755))

	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)
	viper.Set("wowPath.classic", classicInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = filepath.Join(topDir, ".release")
	LinkParams.WSLPathToAddonReleaseDir = releaseDir

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(retailInstall, "Interface", "AddOns", addonName))
	assert.NoFileExists(t, filepath.Join(classicInstall, "Interface", "AddOns", addonName))
}

func TestLinkReplacesExistingSymlinkWithoutForce(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"
	newSource := filepath.Join(releaseDir, addonName)
	require.NoError(t, os.MkdirAll(newSource, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)

	target := filepath.Join(retailInstall, "Interface", "AddOns", addonName)
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
	oldSource := filepath.Join(t.TempDir(), "old-target")
	require.NoError(t, os.MkdirAll(oldSource, 0755))
	require.NoError(t, os.Symlink(oldSource, target))

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	require.NoError(t, Link())

	resolved, err := os.Readlink(target)
	require.NoError(t, err)
	assert.Equal(t, newSource, resolved)
}

func TestLinkRequiresForceForExistingDirectory(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"
	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)

	target := filepath.Join(retailInstall, "Interface", "AddOns", addonName)
	require.NoError(t, os.MkdirAll(target, 0755))

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	err := Link()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use --force to overwrite")
}

func TestLinkForceOverwritesExistingDirectory(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"
	newSource := filepath.Join(releaseDir, addonName)
	require.NoError(t, os.MkdirAll(newSource, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 110007\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)

	target := filepath.Join(retailInstall, "Interface", "AddOns", addonName)
	require.NoError(t, os.MkdirAll(target, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "existing.txt"), []byte("old"), 0644))

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir
	LinkParams.Force = true

	require.NoError(t, Link())
	assertSymlinkExists(t, target)

	resolved, err := os.Readlink(target)
	require.NoError(t, err)
	assert.Equal(t, newSource, resolved)
}

func assertSymlinkExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Lstat(path)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink)
}

// Reproduces the reported failure: wowPath.classicBeta is configured, the TOC
// declares a Forever interface, but nothing was linked because viper hands back
// "classicbeta" and the flavor lookup was case-sensitive.
func TestLinkResolvesLowercasedConfigKeys(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"

	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 16001\n\nCore.lua\n"), 0644))

	classicBetaInstall := filepath.Join(t.TempDir(), "classic-beta")
	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(classicBetaInstall, 0755))
	require.NoError(t, os.MkdirAll(retailInstall, 0755))

	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)
	viper.Set("wowPath.classicBeta", classicBetaInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(classicBetaInstall, "Interface", "AddOns", addonName))
	assert.NoFileExists(t, filepath.Join(retailInstall, "Interface", "AddOns", addonName))
}

// A client installed after `config` last ran has no entry of its own, so it is
// derived from the base path.
func TestLinkDiscoversUnconfiguredFlavorFromBasePath(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"

	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 16001\n\nCore.lua\n"), 0644))

	basePath := filepath.Join(t.TempDir(), "wow")
	discovered := filepath.Join(basePath, "_classic_beta_")
	require.NoError(t, os.MkdirAll(discovered, 0755))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))

	viper.Set("wowPath.base", basePath)
	viper.Set("wowPath.retail", retailInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(discovered, "Interface", "AddOns", addonName))
}

// A configured path for a client that is not installed used to abort the whole
// run; it should just be skipped.
func TestLinkSkipsMissingInstallPathWithoutFailing(t *testing.T) {
	defer resetBuildParams()
	defer resetLinkParams()
	viper.Reset()
	defer viper.Reset()

	topDir := t.TempDir()
	releaseDir := filepath.Join(topDir, ".release")
	addonName := "TestAddon"

	require.NoError(t, os.MkdirAll(filepath.Join(releaseDir, addonName), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(topDir, addonName+".toc"), []byte("## Interface: 16001, 120100\n\nCore.lua\n"), 0644))

	retailInstall := filepath.Join(t.TempDir(), "retail")
	require.NoError(t, os.MkdirAll(retailInstall, 0755))
	missingInstall := filepath.Join(t.TempDir(), "not-installed")

	viper.Set("wowPath.base", filepath.Join(t.TempDir(), "wow"))
	viper.Set("wowPath.retail", retailInstall)
	viper.Set("wowPath.classicBeta", missingInstall)

	BuildParams.TopDir = topDir
	BuildParams.ReleaseDir = releaseDir

	require.NoError(t, Link())

	assertSymlinkExists(t, filepath.Join(retailInstall, "Interface", "AddOns", addonName))
	assert.NoDirExists(t, missingInstall)
}
