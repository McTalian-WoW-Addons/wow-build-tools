package upload

import (
	"fmt"
	"os"
	"strings"

	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

type UploadArgs struct {
	Input             string
	Label             string
	InterfaceVersions []int
	Changelog         string
	ReleaseType       string
}

var UploadParams = &UploadArgs{}

type preparePayload struct {
	toc       *toc.Toc
	changelog *changelog.Changelog
}

func prepareUpload() (prepPayload *preparePayload, err error) {
	defer func() {
		if err != nil {
			logger.Error("Failed to prepare upload: %v", err)
		}
	}()

	tmp := os.TempDir()
	tmpToc, err := os.CreateTemp(tmp, "wbt*.toc")
	if err != nil {
		return
	}
	defer func() {
		_ = tmpToc.Close()
		_ = os.Remove(tmpToc.Name())
	}()

	changelogPath := UploadParams.Changelog
	if UploadParams.Changelog == "" {
		var tmpChangelog *os.File
		tmpChangelog, err = os.CreateTemp(tmp, "wbtChangelog*.md")
		if err != nil {
			return
		}
		defer func() {
			_ = tmpChangelog.Close()
			_ = os.Remove(tmpChangelog.Name())
		}()

		_, err = tmpChangelog.WriteString("No changelog provided")
		if err != nil {
			return
		}
		err = tmpChangelog.Sync()
		if err != nil {
			return
		}

		changelogPath = tmpChangelog.Name()
	}

	cLog := &changelog.Changelog{
		PreExistingFilePath: changelogPath,
		MarkupType:          changelog.MarkdownMT,
	}

	interfaceStringList := []string{}
	for _, i := range UploadParams.InterfaceVersions {
		interfaceStringList = append(interfaceStringList, fmt.Sprintf("%d", i))
	}

	interfaceString := strings.Join(interfaceStringList, ",")
	_, err = fmt.Fprintf(tmpToc, "## Interface: %s", interfaceString)
	if err != nil {
		return
	}
	err = tmpToc.Sync()
	if err != nil {
		return
	}

	tocFile, err := toc.NewToc(tmpToc.Name())
	if err != nil {
		return
	}

	prepPayload = &preparePayload{
		toc:       tocFile,
		changelog: cLog,
	}
	return
}
