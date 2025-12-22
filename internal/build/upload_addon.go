package build

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/pkg"
	"github.com/McTalian/wow-build-tools/internal/repo"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/McTalian/wow-build-tools/internal/upload"
)

type uploadToDistrosArgs struct {
	zipFilePath    string
	fileLabel      string
	tocFiles       []*toc.Toc
	pkgMeta        *pkg.PkgMeta
	cl             *changelog.Changelog
	releaseDir     string
	releaseType    string
	projectName    string
	projectVersion string
	vR             repo.VcsRepo
	isNoLib        bool
	noLibFileName  string
}

func uploadToDistros(args uploadToDistrosArgs) (err error) {
	defer func() {
		if err != nil {
			l.Error("Error uploading to distributions: %v", err)
		}
	}()

	zipFilePath := args.zipFilePath
	fileLabel := args.fileLabel
	tocFiles := args.tocFiles
	pkgMeta := args.pkgMeta
	cl := args.cl
	releaseDir := args.releaseDir
	releaseType := args.releaseType
	projectName := args.projectName
	projectVersion := args.projectVersion
	vR := args.vR
	isNoLib := args.isNoLib
	noLibFileName := args.noLibFileName

	uploadsToAttempt := 4
	var uploadWGroup sync.WaitGroup
	uploadErrChan := make(chan error, uploadsToAttempt)
	uploadWGroup.Add(uploadsToAttempt)

	go func() {
		defer uploadWGroup.Done()
		curseArgs := upload.UploadCurseArgs{
			ZipPath:     zipFilePath,
			FileLabel:   fileLabel,
			TocFiles:    tocFiles,
			PkgMeta:     pkgMeta,
			Changelog:   cl,
			ReleaseType: releaseType,
		}
		if err = upload.UploadToCurse(curseArgs); err != nil {
			l.Error("Curse Upload Error: %v", err)
			uploadErrChan <- err
			return
		}
	}()

	go func() {
		defer uploadWGroup.Done()
		wowiArgs := upload.UploadWowiArgs{
			TocFiles:       tocFiles,
			ProjectVersion: projectVersion,
			ZipPath:        zipFilePath,
			FileLabel:      fileLabel,
			Changelog:      cl,
			ReleaseType:    releaseType,
		}
		if err = upload.UploadToWowi(wowiArgs); err != nil {
			l.Error("WoW Interface Upload Error: %v", err)
			uploadErrChan <- err
			return
		}
	}()

	go func() {
		defer uploadWGroup.Done()
		wagoArgs := upload.UploadWagoArgs{
			ZipPath:     zipFilePath,
			FileLabel:   fileLabel,
			TocFiles:    tocFiles,
			Changelog:   cl,
			ReleaseType: releaseType,
		}
		if err = upload.UploadToWago(wagoArgs); err != nil {
			l.Error("Wago Upload Error: %v", err)
			uploadErrChan <- err
			return
		}
	}()

	go func() {
		defer uploadWGroup.Done()
		githubArgs := upload.UploadGitHubArgs{
			ZipPaths:       []string{zipFilePath},
			ProjectName:    projectName,
			ProjectVersion: projectVersion,
			Repo:           vR,
			Changelog:      cl,
			ReleaseType:    releaseType,
		}
		if isNoLib {
			githubArgs.ZipPaths = append(githubArgs.ZipPaths, filepath.Join(releaseDir, noLibFileName+".zip"))
		}

		if err = upload.UploadToGitHub(githubArgs); err != nil {
			l.Error("GitHub Upload Error: %v", err)
			uploadErrChan <- err
			return
		}
	}()

	uploadWGroup.Wait()
	close(uploadErrChan)

	// Collect errors
	errsEncountered := 0
	errStr := ""
	for err := range uploadErrChan {
		if err != nil {
			errsEncountered++
			errStr += fmt.Sprintf("\n  - %s", err.Error())
		}
	}

	if errsEncountered > 0 {
		err = fmt.Errorf("encountered %d errors while uploading to distributions:%s", errsEncountered, errStr)
	}

	return
}
