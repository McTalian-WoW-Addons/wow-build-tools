package toc

import (
	"encoding/json"
	"fmt"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

func RunTocUpdate(outputJson bool) (err error) {
	l := logger.GetSubLog("TOC_UPDATE")

	// When outputting JSON, suppress all non-error logging to keep stdout clean
	if outputJson {
		l.SetLogLevel(logger.ERROR)
	}

	defer func() {
		if err != nil {
			l.Error("TOC Update Error: %v", err)
		}
	}()

	tocFiles, err := FindTocFiles(TocParams.AddonDir)
	if err != nil {
		return
	}

	// Aggregate results from all TOC files
	aggregatedResult := &UpdateResult{}

	for _, tocFilePath := range tocFiles {
		var tocFile *Toc
		tocFile, err = NewToc(tocFilePath)
		if err != nil {
			return
		}

		result, err := tocFile.UpdateInterfaceVersions(FlavorReleaseInfo{
			IsBeta: TocParams.Beta,
			IsTest: TocParams.Ptr,
		})
		if err != nil {
			return err
		}

		// Aggregate results
		if result != nil {
			aggregatedResult.TotalAdded += result.TotalAdded
			aggregatedResult.TotalRemoved += result.TotalRemoved
		}
	}

	// Output JSON if flag set
	if outputJson {
		jsonBytes, err := json.MarshalIndent(aggregatedResult, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %v", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Only print success message when not outputting JSON
		l.Success("TOC file(s) updated successfully")
	}

	return nil
}
