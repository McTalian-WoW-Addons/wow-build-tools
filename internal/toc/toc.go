package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

type Toc struct {
	Filepath  string
	Interface []int
	Title     string
	Notes     string
	Version   string
	Files     []string
	CurseId   string
	WowiId    string
	WagoId    string
	Flavor    GameFlavor
}

func (t *Toc) addGameVersionsFromToc() map[GameFlavor][]string {
	for _, interfaceVersion := range t.Interface {
		// Grab the right-most 2 digits for the patch version
		patchVersion := interfaceVersion % 100
		// Grab the middle 2 digits for the minor version
		minorVersion := (interfaceVersion / 100) % 100
		// Grab the left-most digits for the major version
		majorVersion := interfaceVersion / 10000

		flavor := getFlavorFromMajorVersion(majorVersion)
		AddGameVersion(flavor, fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion))
		AddGameInterface(flavor, interfaceVersion)
	}

	return gameVersions
}

func (t *Toc) GetTocEntriesTree(addonDir string, ignoredFiles []string, l *logger.Logger) (*TocTree, error) {
	var tocTree TocTree
	for _, file := range t.Files {
		if slices.Contains(ignoredFiles, file) {
			l.Verbose("Ignoring file in TOC tree: %s", file)
			continue
		}

		tocEntry := TocEntry{}
		tocEntry.Filepath = filepath.Join(addonDir, file)
		err := tocEntry.populateEntries(ignoredFiles, l)
		if err != nil {
			return nil, err
		}
		tocTree.Entries = append(tocTree.Entries, &tocEntry)
	}

	return &tocTree, nil
}

func NewToc(path string) (*Toc, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading TOC file: %v", err)
	}

	toc, err := parse(path, string(contents))
	if err != nil {
		return nil, fmt.Errorf("error parsing TOC file: %v", err)
	}

	return toc, nil
}
