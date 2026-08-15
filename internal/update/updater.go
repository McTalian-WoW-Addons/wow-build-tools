package update

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// release is a self-update-backend-agnostic view of a detected GitHub
// release. Keeping it separate from any specific library's types means only
// the adapter below needs to change when the underlying self-update library
// changes.
type release struct {
	Version      string
	AssetURL     string
	ReleaseNotes string
}

// selfUpdater abstracts the self-update backend so the control flow in
// updatebin.go can be unit tested independently of, and later swapped to a
// different, self-update library without touching that control flow.
type selfUpdater interface {
	DetectLatest(ctx context.Context, repo string) (*release, bool, error)
	UpdateTo(ctx context.Context, rel *release, cmdPath string) error
}

// rhysdUpdater adapts github.com/rhysd/go-github-selfupdate to selfUpdater.
// rhysd/go-github-selfupdate depends on blang/semver internally (its
// Release.Version is a semver.Version) — we only call .String() on it here,
// never import blang/semver ourselves. Every other file in this package works
// with plain version strings via hashicorp/go-version.
type rhysdUpdater struct{}

func (rhysdUpdater) DetectLatest(_ context.Context, repo string) (*release, bool, error) {
	latest, found, err := selfupdate.DetectLatest(repo)
	if err != nil || !found || latest == nil {
		// Normalize to false whenever release is nil, even if the
		// underlying library claims found=true, so callers can trust
		// !found as the sole "nothing to do" signal.
		return nil, false, err
	}
	return &release{
		Version:      latest.Version.String(),
		AssetURL:     latest.AssetURL,
		ReleaseNotes: latest.ReleaseNotes,
	}, true, nil
}

// UpdateTo mirrors go-github-selfupdate's Updater.UpdateCommand preflight:
// if cmdPath is a symlink, resolve it first so the update overwrites the
// real binary rather than replacing the symlink itself.
func (rhysdUpdater) UpdateTo(_ context.Context, rel *release, cmdPath string) error {
	if stat, err := os.Lstat(cmdPath); err == nil && stat.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(cmdPath); err == nil {
			cmdPath = resolved
		}
	}
	return selfupdate.UpdateTo(rel.AssetURL, cmdPath)
}
