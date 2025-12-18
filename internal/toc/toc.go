package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

func (t *Toc) UpdateInterfaceVersions(FlavorReleaseInfo FlavorReleaseInfo) error {
	availableInterfaces, err := CheckForInterfaceBumps(FlavorReleaseInfo)
	if err != nil {
		return fmt.Errorf("error checking for interface bumps: %v", err)
	}

	// Update the toc file with the new interface versions
	contents, err := os.ReadFile(t.Filepath)
	if err != nil {
		return fmt.Errorf("error reading TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## Interface:") {
			var interfaceStrings []string
			for _, iface := range availableInterfaces {
				interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
			}
			newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
			lines[i] = newInterfaceLine
			break
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(t.Filepath, []byte(newContents), 0644)
	if err != nil {
		return fmt.Errorf("error writing updated TOC file: %v", err)
	}

	return nil
}

func (t *Toc) GetFlavorsFromInterfaces() []GameFlavor {
	var flavors []GameFlavor
	flavorSet := make(map[GameFlavor]bool)

	for _, interfaceVersion := range t.Interface {
		majorVersion := interfaceVersion / 10000
		flavor := getFlavorFromMajorVersion(majorVersion)
		if !flavorSet[flavor] {
			flavors = append(flavors, flavor)
			flavorSet[flavor] = true
		}
	}

	return flavors
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
