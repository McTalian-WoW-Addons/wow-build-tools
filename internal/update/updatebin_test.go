package update

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUpdater is a selfUpdater test double that records every call it
// receives so tests can assert on control flow without a network.
type fakeUpdater struct {
	detectLatestRelease *release
	detectLatestFound   bool
	detectLatestErr     error

	updateToErr error

	detectLatestCalls int
	updateToCalls     int
	updateToRel       *release
}

func (f *fakeUpdater) DetectLatest(_ context.Context, _ string) (*release, bool, error) {
	f.detectLatestCalls++
	return f.detectLatestRelease, f.detectLatestFound, f.detectLatestErr
}

func (f *fakeUpdater) UpdateTo(_ context.Context, rel *release, _ string) error {
	f.updateToCalls++
	f.updateToRel = rel
	return f.updateToErr
}

// testSeams swaps every package-level seam for the duration of a test and
// restores the originals on cleanup, so tests never touch the real network,
// real stdin, or actually exit the test process.
func testSeams(t *testing.T, fu *fakeUpdater, stdinInput string) *bool {
	t.Helper()

	origVersion := currentVersion
	origUpdater := updater
	origStdin := stdin
	origOsExit := osExit
	origExecutablePath := executablePath

	exited := false
	currentVersion = "1.0.0"
	updater = fu
	stdin = strings.NewReader(stdinInput)
	osExit = func(int) { exited = true }
	executablePath = func() (string, error) { return "/fake/path/wow-build-tools", nil }

	t.Cleanup(func() {
		currentVersion = origVersion
		updater = origUpdater
		stdin = origStdin
		osExit = origOsExit
		executablePath = origExecutablePath
	})

	return &exited
}

func TestConfirmAndSelfUpdate(t *testing.T) {
	t.Run("invalid current version skips without calling updater", func(t *testing.T) {
		fu := &fakeUpdater{}
		exited := testSeams(t, fu, "")
		currentVersion = "LOCAL"

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.detectLatestCalls)
		assert.False(t, *exited)
	})

	t.Run("DetectLatest error returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestErr: errors.New("network down")}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("no release found returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: false}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("latest equal to current returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "1.0.0"}}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("latest older than current returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "0.9.0"}}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("unparsable latest version returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "not-a-version"}}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("newer release, user declines with n", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "n\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("newer release, empty input declines", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("newer release, invalid input declines", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "maybe\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("newer release, user accepts, update succeeds and exits", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0", AssetURL: "https://example.com/asset"}}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		require.Equal(t, 1, fu.updateToCalls)
		assert.Equal(t, "https://example.com/asset", fu.updateToRel.AssetURL)
		assert.True(t, *exited)
	})

	t.Run("newer release, user accepts uppercase Y, update succeeds and exits", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "Y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 1, fu.updateToCalls)
		assert.True(t, *exited)
	})

	t.Run("newer release, executablePath error skips update", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "y\n")
		executablePath = func() (string, error) { return "", errors.New("no exe") }

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("newer release, UpdateTo error does not exit", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0"},
			updateToErr:         errors.New("download failed"),
		}
		exited := testSeams(t, fu, "y\n")

		ConfirmAndSelfUpdate()

		assert.Equal(t, 1, fu.updateToCalls)
		assert.False(t, *exited)
	})

	t.Run("stdin read error returns without updating", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := testSeams(t, fu, "")
		stdin = erroringReader{} // overrides the reader testSeams just set up

		ConfirmAndSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, *exited)
	})
}

func TestDoSelfUpdate(t *testing.T) {
	t.Run("invalid current version skips without calling updater", func(t *testing.T) {
		fu := &fakeUpdater{}
		testSeams(t, fu, "")
		currentVersion = "LOCAL"

		DoSelfUpdate()

		assert.Equal(t, 0, fu.detectLatestCalls)
	})

	t.Run("DetectLatest error returns", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestErr: errors.New("network down")}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("no release found is treated as up to date", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: false}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("latest equal to current is up to date", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "1.0.0"}}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("unparsable latest version returns", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "not-a-version"}}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("latest newer than current updates, no prompt", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 1, fu.updateToCalls)
	})

	// DoSelfUpdate mirrors go-github-selfupdate's UpdateCommand, which only
	// guards on equality, not ordering -- unlike ConfirmAndSelfUpdate's LTE
	// guard. This is an existing asymmetry between the two entrypoints, not
	// something this refactor introduced; characterizing it here so a
	// future change can't flip it silently.
	t.Run("latest older than current still updates (equality-only guard)", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "0.9.0"}}
		testSeams(t, fu, "")

		DoSelfUpdate()

		assert.Equal(t, 1, fu.updateToCalls)
	})

	t.Run("executablePath error skips update", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		testSeams(t, fu, "")
		executablePath = func() (string, error) { return "", errors.New("no exe") }

		DoSelfUpdate()

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("UpdateTo error does not panic", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0"},
			updateToErr:         errors.New("download failed"),
		}
		testSeams(t, fu, "")

		assert.NotPanics(t, func() { DoSelfUpdate() })
		assert.Equal(t, 1, fu.updateToCalls)
	})

	t.Run("successful update passes the detected release through", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0", AssetURL: "https://example.com/asset", ReleaseNotes: "notes"},
		}
		testSeams(t, fu, "")

		DoSelfUpdate()

		require.Equal(t, 1, fu.updateToCalls)
		assert.Equal(t, "https://example.com/asset", fu.updateToRel.AssetURL)
		assert.Equal(t, "notes", fu.updateToRel.ReleaseNotes)
	})
}

// erroringReader always returns an error, simulating a broken stdin.
type erroringReader struct{}

func (erroringReader) Read(_ []byte) (int, error) {
	return 0, io.ErrClosedPipe
}
