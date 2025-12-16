package toc

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

type TocEntry struct {
	Filepath string
	Entries  []*TocEntry
}

func (e *TocEntry) populateEntries(ignoredFiles []string, l *logger.Logger) error {
	if filepath.Ext(e.Filepath) == ".xml" {
		files, err := readFilesFromXmlFile(e.Filepath)
		if err != nil {
			return fmt.Errorf("error parsing XML file '%s': %v", e.Filepath, err)
		}

		for _, f := range files {
			if slices.Contains(ignoredFiles, f) {
				if l != nil {
					l.Verbose("Ignoring file in XML: %s", f)
				}
				continue
			}

			var tocEntry TocEntry
			tocEntry.Filepath = f
			err := tocEntry.populateEntries(ignoredFiles, l)
			if err != nil {
				return err
			}

			e.Entries = append(e.Entries, &tocEntry)
		}
	}

	return nil
}
