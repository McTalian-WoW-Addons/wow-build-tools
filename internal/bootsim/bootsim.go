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

// Package bootsim embeds the boot-order smoke test (boot_sim.lua and its CLI
// entrypoint boot_sim_cli.lua) and runs it against a real system Lua
// interpreter, as a subprocess.
//
// An earlier version of this ran on gopher-lua, a Go-native reimplementation
// of the interpreter, embedded directly in-process — no external dependency,
// but its memory overhead per Lua table/value OOM-killed on data-heavy addon
// trees (TokenTransmogTooltips' ~424-file raid data tree hit multiple GB and
// never finished; the same script under real Lua finishes in under a second
// at negligible memory). Shelling out to a real interpreter fixes that and
// matches WoW's own Lua 5.1 semantics exactly, at the cost of requiring
// lua5.1/lua5.4/lua on PATH — the same dependency the busted-based specs in
// this workspace already require, but new for the plain Go binary, and one
// that Windows addon devs in particular are unlikely to already have.
package bootsim

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed boot_sim.lua
var bootSimSource string

//go:embed boot_sim_cli.lua
var bootSimCliSource string

// Checked in order: lua5.1 matches the WoW client exactly, lua5.4 is what
// this workspace's busted suites already run on host, lua is whatever the
// system default happens to be.
var luaCandidates = []string{"lua5.1", "lua5.4", "lua"}

func findLuaInterpreter() (string, error) {
	for _, candidate := range luaCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf(
		"no Lua interpreter found on PATH (tried %v) - install lua5.1 (matches the WoW client) or lua5.4",
		luaCandidates,
	)
}

// Step is one file-load or login-event step the simulator ran.
type Step struct {
	Phase string // "load" or "event"
	Label string // file path for "load", event name for "event"
	OK    bool
	Err   string
}

type simResultStep struct {
	Phase string  `json:"phase"`
	Label string  `json:"label"`
	OK    bool    `json:"ok"`
	Err   *string `json:"err"`
}

type simResult struct {
	OK    bool            `json:"ok"`
	Steps []simResultStep `json:"steps"`
}

// Run resolves tocPath's real .toc/XML load order, loads every file against
// a mocked WoW API, and fires ADDON_LOADED/PLAYER_ENTERING_WORLD — all via a
// subprocess running the embedded boot_sim.lua/boot_sim_cli.lua on a real
// Lua interpreter found on PATH.
//
// mockPath, if non-empty, is a plain Lua file executed before simulation —
// the same role an addon's busted `_mocks/helper.lua` plays, just without
// needing busted. Anything it sets as a real global (e.g. `_G.Enum = {...}`)
// is used as-is; anything it doesn't cover falls back to a silent chainable
// stub. That fallback only fires for globals that don't exist at all — if
// mockPath defines `Enum` as a real (partial) table, indexing a field it
// doesn't cover is a genuine Lua error, same as in the real client, not a
// silently swallowed stub.
func Run(tocPath, addonName, mockPath string) (ok bool, steps []Step, err error) {
	luaBin, err := findLuaInterpreter()
	if err != nil {
		return false, nil, err
	}

	tempDir, err := os.MkdirTemp("", "wow-build-tools-bootsim-*")
	if err != nil {
		return false, nil, fmt.Errorf("failed to create temp dir for boot sim scripts: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if writeErr := os.WriteFile(filepath.Join(tempDir, "boot_sim.lua"), []byte(bootSimSource), 0o644); writeErr != nil {
		return false, nil, fmt.Errorf("failed to write boot_sim.lua: %w", writeErr)
	}

	cliPath := filepath.Join(tempDir, "boot_sim_cli.lua")
	if writeErr := os.WriteFile(cliPath, []byte(bootSimCliSource), 0o644); writeErr != nil {
		return false, nil, fmt.Errorf("failed to write boot_sim_cli.lua: %w", writeErr)
	}

	cmd := exec.Command(luaBin, cliPath, tocPath, addonName, mockPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	// Exit code 2 means boot_sim_cli.lua hit a setup error (bad args, an
	// unreadable mocks file, or a crash before results existed) and already
	// wrote a clear diagnostic to stderr — surface that as-is.
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) && exitErr.ExitCode() == 2 {
		return false, nil, fmt.Errorf("%s", bytes.TrimSpace(stderr.Bytes()))
	}

	// Otherwise stdout should hold exactly one line of JSON (print() during
	// simulation is redirected to stderr by boot_sim_cli.lua, but split on
	// newline defensively rather than trusting that always holds).
	lines := bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n"))
	lastLine := lines[len(lines)-1]

	var result simResult
	if jsonErr := json.Unmarshal(lastLine, &result); jsonErr != nil {
		return false, nil, fmt.Errorf(
			"boot simulation did not produce parseable output (run error: %v): %s",
			runErr,
			stderr.String(),
		)
	}

	for _, s := range result.Steps {
		step := Step{Phase: s.Phase, Label: s.Label, OK: s.OK}
		if s.Err != nil {
			step.Err = *s.Err
		}
		steps = append(steps, step)
	}

	return result.OK, steps, nil
}
