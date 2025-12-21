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
	"github.com/McTalian/wow-build-tools/internal/upload"
	"github.com/spf13/cobra"
)

// wagoCmd represents the wago command
var wagoCmd = &cobra.Command{
	Use:   "wago",
	Short: "Upload the specified file to Wago.io",
	Long: `Upload the input zip file to Wago.io.
	
	Input, label, and Wago.io project ID are required.
	The WAGO_API_TOKEN environment variable must also be set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return upload.RunUploadWago()
	},
}

func init() {
	uploadCmd.AddCommand(wagoCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wagoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wagoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	wagoCmd.Flags().StringVarP(&upload.UploadWagoParams.WagoId, "wagoId", "a", "", "Set the Wago project ID for uploading. (Use 0 to unset the TOC value)")
	err := wagoCmd.MarkFlagRequired("wagoId")
	if err != nil {
		panic(err)
	}
}
