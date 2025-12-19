package toc

import (
	"github.com/McTalian/wow-build-tools/internal/logger"
)

func RunTocUpdate() (err error) {
	l := logger.GetSubLog("TOC_UPDATE")

	defer func() {
		if err != nil {
			l.Error("TOC Update Error: %v", err)
		}
	}()

	tocFiles, err := FindTocFiles(TocParams.AddonDir)
	if err != nil {
		return
	}

	for _, tocFilePath := range tocFiles {
		var tocFile *Toc
		tocFile, err = NewToc(tocFilePath)
		if err != nil {
			return
		}

		err = tocFile.UpdateInterfaceVersions(FlavorReleaseInfo{
			IsBeta: TocParams.Beta,
			IsTest: TocParams.Ptr,
		})
		if err != nil {
			return
		}
	}

	l.Success("TOC file(s) updated successfully")
	return nil
}
