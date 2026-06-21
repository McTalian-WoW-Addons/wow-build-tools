package build

import (
	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/pkg"
	"github.com/McTalian/wow-build-tools/internal/repo"
)

func prepareChangeLog(skip bool, vR repo.VcsRepo, pkgMeta *pkg.PkgMeta, changelogTitle, packageDir, topDir string) (cl *changelog.Changelog, err error) {
	defer func() {
		if err != nil {
			l.Error("Changelog Preparation Error: %v", err)
		}
	}()

	cl = &changelog.Changelog{}
	if skip {
		return
	} else {
		cl, err = changelog.NewChangelog(vR, pkgMeta, changelogTitle, packageDir, topDir)
		if err != nil {
			return
		}
		err = cl.GetChangelog()
		if err != nil {
			return
		}
	}

	return
}
