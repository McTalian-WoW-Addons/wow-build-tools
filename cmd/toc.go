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
	"github.com/McTalian/wow-build-tools/internal/toc"

	"github.com/spf13/cobra"
)

// tocCmd represents the toc command
var tocCmd = &cobra.Command{
	Use:   "toc",
	Short: "Tools related to the addon toc file",
	Long:  `Verification tools and utilities for the addon toc file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(tocCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tocCmd.PersistentFlags().String("foo", "", "A help for foo")
	tocCmd.PersistentFlags().StringVarP(&toc.TocParams.AddonDir, "addonDir", "a", ".", "Path to the addon directory (defaults to current working directory)")
	tocCmd.PersistentFlags().StringVarP(&toc.TocParams.AddonDir, "topDir", "t", ".", "Path to the addon directory (defaults to current working directory)")
	if tocCmd.PersistentFlags().MarkDeprecated("topDir", "please use --addonDir instead") != nil {
		panic("failed to mark 'topDir' flag as deprecated")
	}

	tocCmd.PersistentFlags().BoolVarP(&toc.TocParams.Beta, "beta", "b", false, "Include beta versions in the TOC check")
	tocCmd.PersistentFlags().BoolVarP(&toc.TocParams.Ptr, "ptr", "p", false, "Include PTR versions in the TOC check")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tocCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
