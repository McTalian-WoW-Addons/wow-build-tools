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

	"github.com/McTalian/wow-build-tools/internal/toc"
)

// tocUpdateCmd represents the tocUpdate command
var tocUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the interface versions in the addon toc file",
	Long: `Parses the addon toc file and updates the interface versions to the latest available
versions based on the selected release channel(s) (normal, beta, and/or PTR).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return toc.RunTocUpdate(tocUpdateJsonFlag)
	},
}

var tocUpdateJsonFlag bool

func init() {
	tocCmd.AddCommand(tocUpdateCmd)

	// JSON output flag
	tocUpdateCmd.Flags().BoolVar(&tocUpdateJsonFlag, "json", false, "Output update results as JSON")
}
