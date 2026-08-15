package update

import (
	"context"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
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

// rhysdUpdater adapts github.com/rhysd/go-github-selfupdate to selfUpdater.
// rhysd/go-github-selfupdate depends on blang/semver internally (its
// Release.Version is a semver.Version) — we only call .String() on it here,
// never import blang/semver ourselves. Every other file in this package works
// with plain version strings via hashicorp/go-version.
type rhysdUpdater struct{}

func (rhysdUpdater) DetectLatest(_ context.Context, repo string) (*release, bool, error) {
	latest, found, err := selfupdate.DetectLatest(repo)
	if err != nil || !found || latest == nil {
		return nil, found, err
	}
	return &release{
		Version:      latest.Version.String(),
		AssetURL:     latest.AssetURL,
		ReleaseNotes: latest.ReleaseNotes,
	}, true, nil
}

func (rhysdUpdater) UpdateTo(_ context.Context, rel *release, cmdPath string) error {
	return selfupdate.UpdateTo(rel.AssetURL, cmdPath)
}
