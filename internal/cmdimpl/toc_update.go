package cmdimpl

import (
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

func RunTocUpdate() error {
	l := logger.GetSubLog("TOC_UPDATE")
	if RootParams.LevelVerbose {
		l.SetLogLevel(logger.VERBOSE)
	} else if RootParams.LevelDebug {
		l.SetLogLevel(logger.DEBUG)
	} else {
		l.SetLogLevel(logger.INFO)
	}

	tocFiles, err := toc.FindTocFiles(TocParams.AddonDir)
	if err != nil {
		return err
	}

	for _, tocFilePath := range tocFiles {
		tocFile, err := toc.NewToc(tocFilePath)
		if err != nil {
			return err
		}

		err = tocFile.UpdateInterfaceVersions(toc.FlavorReleaseInfo{
			IsBeta: TocParams.Beta,
			IsTest: TocParams.Ptr,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
