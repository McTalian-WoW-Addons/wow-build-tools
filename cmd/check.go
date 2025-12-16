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
	"github.com/spf13/cobra"

	"github.com/McTalian/wow-build-tools/internal/cmdimpl"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for common issues in the addon toc file",
	Long: `Checks for common issues related to the addon toc file.

- Check that the toc file and the addon folder have the same name.
- Check that all of your files are included via the toc file or the tree of its included XML files.
- Check for valid and outdated interface versions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdimpl.TocCheckParams.LevelVerbose = LevelVerbose
		cmdimpl.TocCheckParams.LevelDebug = LevelDebug

		return cmdimpl.RunTocCheck()
	},
}

func init() {
	tocCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	checkCmd.Flags().StringArrayVarP(&cmdimpl.TocCheckParams.IgnoreFiles, "ignore", "x", []string{}, "Files to ignore during the check (if XML are provided, their specified includes will also be ignored)")

	checkCmd.Flags().BoolVarP(&cmdimpl.TocCheckParams.SkipInterfaceCheck, "skip-interface-check", "", false, "Skip checking the interface version")
	checkCmd.Flags().BoolVarP(&cmdimpl.TocCheckParams.SkipMissingFilesCheck, "skip-missing-files-check", "", false, "Skip checking for missing files")
	checkCmd.Flags().BoolVarP(&cmdimpl.TocCheckParams.SkipNameCheck, "skip-name-check", "", false, "Skip checking that the toc file and the addon folder have the same name")
}
