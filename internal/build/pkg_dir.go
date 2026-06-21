package build

import (
	"fmt"

	"github.com/McTalian/wow-build-tools/internal/license"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/pkg"
	"github.com/McTalian/wow-build-tools/internal/repo"
)

func initializePkgDir(topDir string, args *BuildArgs, projectName string, pkgMeta *pkg.PkgMeta, vR repo.VcsRepo) (packageDir string, err error) {
	defer func() {
		if err != nil {
			l.Error("Error initializing package directory: %v", err)
		}
	}()

	copyLogGroup := logger.NewLogGroup(fmt.Sprintf("%sPreparing Package Directory", logger.Package))
	defer copyLogGroup.Flush(true)

	l.Debug("Top Directory: %s", topDir)
	l.Debug("Top Directory from Flags: %s", args.TopDir)
	l.Debug("Release Directory: %s", args.ReleaseDir)
	l.Debug("Project Name: %s", projectName)
	packageDir, err = pkg.PreparePkgDir(projectName, args.ReleaseDir, args.KeepPackageDir)
	l.Debug("Package Directory: %s", packageDir)
	if err != nil {
		return
	}

	err = license.EnsureLicensePresent(pkgMeta.License, topDir, packageDir, args.CurseId)
	if err != nil {
		return
	}

	if !args.SkipCopy {
		projCopy := pkg.NewPkgCopy(topDir, packageDir, pkgMeta.Ignore, vR)
		err = projCopy.CopyToPackageDir(copyLogGroup)
		if err != nil {
			return
		}
	}

	return
}
