package build

import (
	"fmt"

	"github.com/McTalian/wow-build-tools/internal/pkg"
)

func parsePkgmeta(currProjName string, topDir string, pkgmetaFile string) (projectName string, pkgMeta *pkg.PkgMeta, err error) {
	defer func() {
		if err != nil {
			l.Error("Pkgmeta Error: %v", err)
		}
	}()

	projectName = currProjName

	parseArgs := pkg.ParseArgs{
		PkgmetaFile: pkgmetaFile,
		PkgDir:      topDir,
	}
	pkgMeta, err = pkg.Parse(&parseArgs)
	if err != nil {
		return
	}

	if pkgMeta.PackageAs != "" && projectName != pkgMeta.PackageAs {
		err = fmt.Errorf("project name (%s) from TOC filename(s) does not match `package-as` name in pkgmeta file (%s)", projectName, pkgMeta.PackageAs)
		return
	}

	l.Verbose("%s", pkgMeta.String())

	if pkgMeta.PackageAs != "" {
		projectName = pkgMeta.PackageAs
	}

	return
}
