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
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/McTalian/wow-build-tools/internal/build"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run build when files change",
	Long: dedent.Dedent(`
	Watches the current directory for changes and runs the build command when a change is detected.
	
	Running "wow-build-tools link" before running this command is recommended to ensure that the build output directories are symlinked to your WoW installation directories.
	
	You can enable "--copyToWowDirs" as an alterative. The build output directories will then be copied to configured WoW installation directories.
	When copying from WSL to the host system, the copies can be slower than desired.
	`),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a context that cancels on interrupt signals (Ctrl+C)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupt signals for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		return build.WatchBuild(ctx)
	},
}

func init() {
	buildCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	watchCmd.Flags().BoolVarP(&build.WatchParams.CopyToWowDirs, "copyToWowDirs", "w", false, "Copy output to configured WoW directories.")
}
