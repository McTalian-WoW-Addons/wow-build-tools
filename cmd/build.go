/*
Copyright © 2025 Rob "McTalian" Anderson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/McTalian/wow-build-tools/internal/build"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a World of Warcraft addon",
	Long:  `This command packages the addon as specified via a pkgmeta file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return build.Build(build.BuildParams)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("topDir") && !cmd.Flags().Changed("releaseDir") {
			build.BuildParams.ReleaseDir = filepath.Join(build.BuildParams.TopDir, ".release")
		}
		return nil
	}

	buildCmd.Flags().SortFlags = false

	buildCmd.PersistentFlags().StringVarP(&build.BuildParams.TopDir, "topDir", "t", ".", "The top level directory of the addon")
	buildCmd.PersistentFlags().StringVarP(&build.BuildParams.ReleaseDir, "releaseDir", "r", build.BuildParams.TopDir+string(os.PathSeparator)+".release", "The directory to output the release files.")
	buildCmd.Flags().StringVarP(&build.BuildParams.PkgmetaFile, "pkgmetaFile", "m", "", "Set the pkgmeta file to use. (Defaults to {topDir}/pkgmeta.yml, {topDir}/pkgmeta.yaml, or {topDir}/.pkgmeta if one exists.)")
	buildCmd.Flags().BoolVarP(&build.BuildParams.KeepPackageDir, "keepPackageDir", "o", false, "Keep existing package directory, overwriting its contents.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.CreateNoLib, "createNoLib", "s", false, "Create a stripped-down \"nolib\" package.")
	buildCmd.Flags().StringVarP(&build.BuildParams.CurseId, "curseId", "p", "", "Set the CurseForge project ID for localization and uploading. (Use 0 to unset the TOC value)")
	buildCmd.Flags().StringVarP(&build.BuildParams.WowiId, "wowiId", "w", "", "Set the WoWInterface project ID for uploading. (Use 0 to unset the TOC value)")
	buildCmd.Flags().StringVarP(&build.BuildParams.WagoId, "wagoId", "a", "", "Set the Wago project ID for uploading. (Use 0 to unset the TOC value)")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SkipCopy, "skipCopy", "c", false, "Skip copying the files to the output directory.")
	buildCmd.Flags().BoolVar(&build.BuildParams.SkipChangelog, "skipChangelog", false, "Skip changelog generation.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SkipExternals, "skipExternals", "e", false, "Skip fetching externals.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.ForceExternals, "forceExternals", "E", false, "Force fetching externals, bypassing the cache.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SkipZip, "skipZip", "z", false, "Skip zipping the package (and uploading).")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SkipUpload, "skipUpload", "d", false, "Skip uploading.")
	buildCmd.Flags().StringVarP(&build.BuildParams.NameTemplate, "nameTemplate", "n", "", "Set the name template to use for the release file. Use \"-n help\" for more info.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SkipLocalization, "skipLocalization", "l", false, "Skip @localization@ keyword replacement.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.OnlyLocalization, "onlyLocalization", "L", false, "Only do @localization@ keyword replacement (skip upload to CurseForge).")
	buildCmd.Flags().BoolVarP(&build.BuildParams.SplitToc, "splitToc", "S", false, "Create a package supporting multiple game types from a single TOC file.")
	buildCmd.Flags().BoolVarP(&build.BuildParams.UnixLineEndings, "unixLineEndings", "u", false, "Use Unix line endings in TOC and XML files.")
	buildCmd.Flags().StringVarP(&build.BuildParams.GameVersion, "gameVersion", "g", "", "Set the game version to use for uploading.")
}
