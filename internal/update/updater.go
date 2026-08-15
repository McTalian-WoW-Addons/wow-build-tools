package update

import (
	"context"
	"os"
	"path/filepath"

	"github.com/creativeprojects/go-selfupdate"
)

// release is a self-update-backend-agnostic view of a detected GitHub
// release. Keeping it separate from any specific library's types means only
// the adapter below needs to change when the underlying self-update library
// changes.
type release struct {
	Version      string
	AssetURL     string
	AssetName    string
	ReleaseNotes string
}

// selfUpdater abstracts the self-update backend so the control flow in
// updatebin.go can be unit tested independently of, and later swapped to a
// different, self-update library without touching that control flow.
type selfUpdater interface {
	DetectLatest(ctx context.Context, repo string) (*release, bool, error)
	UpdateTo(ctx context.Context, rel *release, cmdPath string) error
}

// creativeprojectsUpdater adapts github.com/creativeprojects/go-selfupdate to
// selfUpdater. It uses the package-level DetectLatest/UpdateTo, the same
// granularity github.com/rhysd/go-github-selfupdate offered: a plain HTTP GET
// of the release asset by URL, no GitHub API auth required for update itself
// (DetectLatest still hits the API, honoring $GITHUB_TOKEN if set). Internally
// this library parses tags with Masterminds/semver/v3, whose Version.String()
// strips any leading "v" the same way blang/semver did — confirmed via probe,
// see internal/update probes in git history.
type creativeprojectsUpdater struct{}

func (creativeprojectsUpdater) DetectLatest(ctx context.Context, repo string) (*release, bool, error) {
	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(repo))
	if err != nil || !found || latest == nil {
		// Normalize to false whenever release is nil, even if the
		// underlying library claims found=true, so callers can trust
		// !found as the sole "nothing to do" signal.
		return nil, false, err
	}
	return &release{
		Version:      latest.Version(),
		AssetURL:     latest.AssetURL,
		AssetName:    latest.AssetName,
		ReleaseNotes: latest.ReleaseNotes,
	}, true, nil
}

// UpdateTo mirrors go-github-selfupdate's Updater.UpdateCommand preflight:
// if cmdPath is a symlink, resolve it first so the update overwrites the
// real binary rather than replacing the symlink itself.
func (creativeprojectsUpdater) UpdateTo(ctx context.Context, rel *release, cmdPath string) error {
	if stat, err := os.Lstat(cmdPath); err == nil && stat.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(cmdPath); err == nil {
			cmdPath = resolved
		}
	}
	return selfupdate.UpdateTo(ctx, rel.AssetURL, rel.AssetName, cmdPath)
}
