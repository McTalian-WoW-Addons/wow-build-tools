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

func fakeExecutablePath() (string, error) {
	return "/fake/path/wow-build-tools", nil
}

func failingExecutablePath() (string, error) {
	return "", errors.New("no exe")
}

func TestConfirmAndSelfUpdate(t *testing.T) {
	t.Run("invalid current version skips without calling updater", func(t *testing.T) {
		fu := &fakeUpdater{}
		exited := false

		confirmAndSelfUpdate(fu, "LOCAL", strings.NewReader(""), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.detectLatestCalls)
		assert.False(t, exited)
	})

	t.Run("DetectLatest error returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestErr: errors.New("network down")}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("no release found returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: false}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("latest equal to current returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "1.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("latest older than current returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "0.9.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("unparsable latest version returns without prompting", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "not-a-version"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("newer release, user declines with n", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("n\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("newer release, empty input declines", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("newer release, invalid input declines", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("maybe\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("newer release, user accepts, update succeeds and exits", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0", AssetURL: "https://example.com/asset"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		require.Equal(t, 1, fu.updateToCalls)
		assert.Equal(t, "https://example.com/asset", fu.updateToRel.AssetURL)
		assert.True(t, exited)
	})

	t.Run("newer release, user accepts uppercase Y, update succeeds and exits", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("Y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 1, fu.updateToCalls)
		assert.True(t, exited)
	})

	t.Run("newer release, executablePath error skips update", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, failingExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("newer release, UpdateTo error does not exit", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0"},
			updateToErr:         errors.New("download failed"),
		}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", strings.NewReader("y\n"), func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 1, fu.updateToCalls)
		assert.False(t, exited)
	})

	t.Run("stdin read error returns without updating", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}
		exited := false

		confirmAndSelfUpdate(fu, "1.0.0", erroringReader{}, func(int) { exited = true }, fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
		assert.False(t, exited)
	})
}

func TestDoSelfUpdate(t *testing.T) {
	t.Run("invalid current version skips without calling updater", func(t *testing.T) {
		fu := &fakeUpdater{}

		doSelfUpdate(fu, "LOCAL", fakeExecutablePath)

		assert.Equal(t, 0, fu.detectLatestCalls)
	})

	t.Run("DetectLatest error returns", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestErr: errors.New("network down")}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("no release found is treated as up to date", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: false}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("latest equal to current is up to date", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "1.0.0"}}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("unparsable latest version returns", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "not-a-version"}}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("latest newer than current updates, no prompt", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 1, fu.updateToCalls)
	})

	// DoSelfUpdate mirrors go-github-selfupdate's UpdateCommand, which only
	// guards on equality, not ordering -- unlike ConfirmAndSelfUpdate's LTE
	// guard. This is an existing asymmetry between the two entrypoints, not
	// something this refactor introduced; characterizing it here so a
	// future change can't flip it silently.
	t.Run("latest older than current still updates (equality-only guard)", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "0.9.0"}}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

		assert.Equal(t, 1, fu.updateToCalls)
	})

	t.Run("executablePath error skips update", func(t *testing.T) {
		fu := &fakeUpdater{detectLatestFound: true, detectLatestRelease: &release{Version: "2.0.0"}}

		doSelfUpdate(fu, "1.0.0", failingExecutablePath)

		assert.Equal(t, 0, fu.updateToCalls)
	})

	t.Run("UpdateTo error does not panic", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0"},
			updateToErr:         errors.New("download failed"),
		}

		assert.NotPanics(t, func() { doSelfUpdate(fu, "1.0.0", fakeExecutablePath) })
		assert.Equal(t, 1, fu.updateToCalls)
	})

	t.Run("successful update passes the detected release through", func(t *testing.T) {
		fu := &fakeUpdater{
			detectLatestFound:   true,
			detectLatestRelease: &release{Version: "2.0.0", AssetURL: "https://example.com/asset", ReleaseNotes: "notes"},
		}

		doSelfUpdate(fu, "1.0.0", fakeExecutablePath)

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
