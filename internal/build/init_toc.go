package build

import (
	"fmt"

	"github.com/McTalian/wow-build-tools/internal/toc"
)

func getTocFiles(topDir string) (projectName string, tocFiles []*toc.Toc, err error) {
	tocFilePaths, err := toc.FindTocFiles(topDir)
	if err != nil {
		return
	}

	projectName = toc.DetermineProjectName(tocFilePaths)

	for _, tocFilePath := range tocFilePaths {
		var t *toc.Toc
		t, err = toc.NewToc(tocFilePath)
		if err != nil {
			return
		}
		tocFiles = append(tocFiles, t)
	}

	if len(tocFiles) == 0 {
		err = fmt.Errorf("no TOC files found in %s", topDir)
		return
	}

	return
}
