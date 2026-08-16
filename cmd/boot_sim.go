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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/McTalian/wow-build-tools/internal/bootsim"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

var bootSimTopDir string
var bootSimMockPath string

var bootSimCmd = &cobra.Command{
	Use:   "boot-sim",
	Short: "Simulate a client login to catch Lua errors before they reach a player",
	Long: `Resolves each .toc file's real load order (following XML Script/Include
chains), loads every file through a real Lua interpreter, and fires the
ADDON_LOADED and PLAYER_ENTERING_WORLD events the same way a real client
does on login.

A curated set of known WoW API globals (see KNOWN_WOW_APIS in boot_sim.lua)
silently resolves to a no-op stub rather than erroring; everything else is
real nil, including addon/library-defined globals like LibStub — that's
deliberate, so common init-guard idioms ("if not LibStub then ...") behave
correctly instead of reading as already loaded. So by default this catches
bugs in the addon's own code — nil references, load-order mistakes, syntax
errors — not general WoW API coverage gaps. Pass --mocks with a plain Lua
file (e.g. a busted _mocks/helper.lua, if it has no busted-specific
dependencies) to seed real WoW API globals first; anything it sets is used
as-is. It runs against every .toc file found (one per game flavor).

Requires lua5.1, lua5.4, or lua on PATH (checked in that order) — the
simulator script is embedded in this binary, but it needs a real
interpreter to run on. lua5.1 matches the WoW client exactly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.GetSubLog("BOOT_SIM")

		tocFiles, err := toc.FindTocFiles(bootSimTopDir)
		if err != nil {
			return err
		}

		addonName := toc.DetermineProjectName(tocFiles)

		anyFailed := false
		for _, tocFile := range tocFiles {
			ok, steps, err := bootsim.Run(tocFile, addonName, bootSimMockPath)
			if err != nil {
				anyFailed = true
				l.Error("%s: %v", tocFile, err)
				continue
			}

			if !ok {
				anyFailed = true
				for _, step := range steps {
					if !step.OK {
						l.Error("%s [%s] %s: %s", tocFile, step.Phase, step.Label, step.Err)
					}
				}
				continue
			}

			l.Info("%s: PASS (%d steps)", tocFile, len(steps))
		}

		if anyFailed {
			return fmt.Errorf("boot simulation failed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bootSimCmd)

	bootSimCmd.Flags().StringVarP(&bootSimTopDir, "topDir", "t", ".", "The top level directory of the addon")
	bootSimCmd.Flags().StringVarP(&bootSimMockPath, "mocks", "m", "", "Path to a Lua file that seeds real WoW API globals before simulation")
}
