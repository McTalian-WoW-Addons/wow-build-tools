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
	"github.com/McTalian/wow-build-tools/internal/build"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
)

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Create symlinks in World of Warcraft AddOns directory to the addon(s) in the build output directory",
	Long: dedent.Dedent(`
		Create symlinks in the World of Warcraft AddOns directory to the addon(s) in the build output directory.
		
		By default, the release directory is assumed to be a ".release" directory in the top level directory of the addon.
		
		If you are developing in WSL, you will need to run this command in Windows in an elevated command prompt.
		You will also need to provide the path to the addon release directory in WSL using the --wsl-path-to-addon-release-dir flag.
		From WSL, run "wslpath -w <path_to_your_releasedir>" to get the Windows path to your release directory.`),
	RunE: func(cmd *cobra.Command, args []string) error {
		return build.Link()
	},
}

func init() {
	buildCmd.AddCommand(linkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// linkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// linkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	linkCmd.Flags().StringVarP(&build.LinkParams.WSLPathToAddonReleaseDir, "wsl-path-to-addon-release-dir", "w", "", "Path to the addon release directory in WSL")
	linkCmd.Flags().BoolVarP(&build.LinkParams.Force, "force", "f", false, "Force linking even if the destination exists (will overwrite)")
	linkCmd.Flags().StringSliceVar(&build.LinkParams.OnlyFlavors, "flavor", []string{}, "Only create links in the specified flavor installations")
}
