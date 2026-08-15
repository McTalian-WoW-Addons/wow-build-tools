package update

import (
	"context"
	"io"
	"os"

	goversion "github.com/hashicorp/go-version"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

// version is replaced by the tag version in the release binary via a literal
// sed match on this exact declaration — see the "Replace version token in
// source" step in .github/workflows/wbt-release-published.yml. Do not
// reformat or rename this line without updating that sed pattern too.
const version = "LOCAL"
const repo = "McTalian-WoW-Addons/wow-build-tools"

// CurrentVersion returns the version this binary was built with -- "LOCAL"
// for a dev build, or the release tag baked in via the sed replacement
// described above.
func CurrentVersion() string {
	return version
}

func ConfirmAndSelfUpdate() {
	confirmAndSelfUpdate(creativeprojectsUpdater{}, version, os.Stdin, os.Exit, os.Executable)
}

func DoSelfUpdate() {
	doSelfUpdate(creativeprojectsUpdater{}, version, os.Executable)
}

func parseVersion(currentVersion string) (*goversion.Version, error) {
	v, err := goversion.NewVersion(currentVersion)
	if err != nil {
		logger.Debug("Running in local, alpha, or beta mode (%s). Skipping self-update.", currentVersion)
		return nil, err
	}
	return v, nil
}

func confirmAndSelfUpdate(u selfUpdater, currentVersion string, stdin io.Reader, osExit func(int), executablePath func() (string, error)) {
	v, err := parseVersion(currentVersion)
	if err != nil {
		return
	}

	latest, found, err := u.DetectLatest(context.Background(), repo)
	if err != nil {
		logger.Error("Error occurred while detecting version: %v", err)
		return
	}
	if !found {
		logger.Debug("Current version is the latest")
		return
	}
	latestVersion, err := goversion.NewVersion(latest.Version)
	if err != nil {
		logger.Error("Error occurred while parsing latest version: %v", err)
		return
	}
	if latestVersion.LessThanOrEqual(v) {
		logger.Debug("Current version is the latest")
		return
	}

	update, err := logger.PromptYesNo(stdin, "Do you want to update to %s? (y/N): ", latest.Version)
	if err != nil {
		logger.Error("Invalid input: %v", err)
		return
	}
	if !update {
		logger.Info("Skipping update, if you change your mind, run `wow-build-tools update` at any time")
		return
	}

	exe, err := executablePath()
	if err != nil {
		logger.Error("Could not locate executable path: %v", err)
		return
	}
	if err := u.UpdateTo(context.Background(), latest, exe); err != nil {
		logger.Error("Error occurred while updating binary: %v", err)
		return
	}

	logger.Success("Successfully updated to version %s", latest.Version)
	logger.Info("Re-run the command to use the new version")
	osExit(0)
}

func doSelfUpdate(u selfUpdater, currentVersion string, executablePath func() (string, error)) {
	v, err := parseVersion(currentVersion)
	if err != nil {
		return
	}

	logger.Info("Checking for newer versions that %s...", v.String())

	latest, found, err := u.DetectLatest(context.Background(), repo)
	if err != nil {
		logger.Error("Binary update failed: %v", err)
		return
	}
	if !found {
		// Mirrors go-github-selfupdate's UpdateCommand: no release detected
		// is treated as already up to date, not an error.
		logger.Info("Current binary is the latest version %s", v.String())
		return
	}
	latestVersion, err := goversion.NewVersion(latest.Version)
	if err != nil {
		logger.Error("Binary update failed: %v", err)
		return
	}
	if latestVersion.Equal(v) {
		logger.Info("Current binary is the latest version %s", v.String())
		return
	}

	exe, err := executablePath()
	if err != nil {
		logger.Error("Binary update failed: %v", err)
		return
	}
	if err := u.UpdateTo(context.Background(), latest, exe); err != nil {
		logger.Error("Binary update failed: %v", err)
		return
	}

	logger.Info("Successfully updated to version %s", latest.Version)
	logger.Info("Release note:\n%s", latest.ReleaseNotes)
}
