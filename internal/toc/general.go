package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

func TocFileToGameFlavor(noExt string) (flavor GameFlavor, suffix string) {
	if strings.Contains(noExt, "-") {
		postDash := strings.Split(noExt, "-")
		if len(postDash) > 1 {
			suffix = postDash[len(postDash)-1]
		}
	} else if strings.Contains(noExt, "_") {
		postUnderscore := strings.Split(noExt, "_")
		if len(postUnderscore) > 1 {
			suffix = postUnderscore[len(postUnderscore)-1]
		}
	}

	normalSuffix := strings.ToLower(suffix)

	switch normalSuffix {
	case "vanilla":
		flavor = ClassicEra
	case "classic":
		flavor = CurrentClassic // Classic progression server (currently Mists)
	case "tbc", "bcc":
		flavor = TbcClassic
	case "wrath", "wotlk", "wotlkc":
		flavor = WotlkClassic
	case "cata":
		flavor = CataClassic
	case "mists":
		flavor = MistsClassic
	case "wod":
		flavor = WodClassic
	case "legion":
		flavor = LegionClassic
	case "bfa":
		flavor = BfaClassic
	case "sl":
		flavor = SlClassic
	case "df":
		flavor = DfClassic
	case "", "mainline":
		flavor = Retail
	default:
		flavor = Unknown
	}

	return
}

func FindTocFiles(path string) ([]string, error) {
	tocFiles := []string{}
	matches, err := filepath.Glob(path + string(os.PathSeparator) + "*.toc")
	if err != nil {
		return tocFiles, fmt.Errorf("error finding TOC file in %s: %v", path, err)
	}

	if len(matches) == 0 {
		return tocFiles, fmt.Errorf("no TOC file found in %s", path)
	}

	tocFiles = append(tocFiles, matches...)

	slices.Sort(tocFiles)

	return tocFiles, nil
}

func DetermineProjectName(tocFiles []string) string {
	projectName := ""
	for _, tocFile := range tocFiles {
		tocFilePath := filepath.Base(tocFile)
		var flavor GameFlavor
		noExt := strings.TrimSuffix(tocFilePath, filepath.Ext(tocFilePath))

		if !strings.Contains(noExt, "-") && !strings.Contains(noExt, "_") {
			projectName = noExt

			break
		}

		flavor, suffix := TocFileToGameFlavor(noExt)
		if flavor != Unknown {
			projectName = strings.ReplaceAll(noExt, "_"+suffix, "")
			projectName = strings.ReplaceAll(projectName, "-"+suffix, "")
			break
		}

	}

	return projectName
}

func parse(filePath, tocContents string) (*Toc, error) {
	toc := &Toc{}
	toc.Filepath = filePath
	baseFilename := filepath.Base(filePath)
	toc.Flavor, _ = TocFileToGameFlavor(strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename)))
	lines := strings.Split(tocContents, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Interface:") && !strings.Contains(line, "## Interface-") {
			// Parse unsuffixed Interface lines (e.g., "## Interface: 110007" or "## Interface: 11505, 50000, 110007")
			interfaceLine := strings.TrimPrefix(line, "## Interface:")
			interfaceLine = strings.TrimSpace(interfaceLine)
			interfaceValues := strings.Split(interfaceLine, ",")
			for _, interfaceValue := range interfaceValues {
				interfaceValue = strings.TrimSpace(interfaceValue)
				interfaceVersion, err := strconv.Atoi(interfaceValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing Interface version: %v", err)
				}
				toc.Interface = append(toc.Interface, interfaceVersion)
			}
		} else if strings.HasPrefix(line, "## Interface-") {
			// Parse suffixed Interface lines (e.g., "## Interface-Vanilla: 11505")
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				interfaceValue := strings.TrimSpace(parts[1])
				interfaceVersion, err := strconv.Atoi(interfaceValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing suffixed Interface version: %v", err)
				}
				toc.Interface = append(toc.Interface, interfaceVersion)
			}
		} else if strings.HasPrefix(line, "## Title:") {
			toc.Title = strings.TrimPrefix(line, "## Title:")
			toc.Title = strings.TrimSpace(toc.Title)
		} else if strings.HasPrefix(line, "## Notes:") {
			toc.Notes = strings.TrimPrefix(line, "## Notes:")
			toc.Notes = strings.TrimSpace(toc.Notes)
		} else if strings.HasPrefix(line, "## Version:") {
			toc.Version = strings.TrimPrefix(line, "## Version:")
			toc.Version = strings.TrimSpace(toc.Version)
		} else if !strings.HasPrefix(line, "#") {
			file := strings.TrimSpace(line)
			if file == "" {
				continue
			}
			toc.Files = append(toc.Files, file)
		} else if strings.HasPrefix(line, "## X-Curse-Project-ID:") {
			toc.CurseId = strings.TrimPrefix(line, "## X-Curse-Project-ID:")
			toc.CurseId = strings.TrimSpace(toc.CurseId)
		} else if strings.HasPrefix(line, "## X-WoWI-ID:") {
			toc.WowiId = strings.TrimPrefix(line, "## X-WoWI-ID:")
			toc.WowiId = strings.TrimSpace(toc.WowiId)
		} else if strings.HasPrefix(line, "## X-Wago-ID:") {
			toc.WagoId = strings.TrimPrefix(line, "## X-Wago-ID:")
			toc.WagoId = strings.TrimSpace(toc.WagoId)
		}
	}

	toc.tocSpecificInterfaces = make(map[GameFlavor][]int)

	toc.addGameVersionsFromToc()

	return toc, nil
}

func GetTocFileTree(path string) ([]string, error) {
	tocFiles, err := FindTocFiles(path)
	if err != nil {
		return nil, fmt.Errorf("error finding TOC files: %v", err)
	}

	var coveredFilesSet = make(map[string]bool)
	for _, tocFile := range tocFiles {
		toc, err := NewToc(tocFile)
		if err != nil {
			return nil, fmt.Errorf("error creating TOC object: %v", err)
		}
		for _, file := range toc.Files {
			coveredFilesSet[filepath.Join(path, file)] = true
		}
	}

	var coveredFiles []string
	for file := range coveredFilesSet {
		coveredFiles = append(coveredFiles, file)
	}

	slices.Sort(coveredFiles)

	return coveredFiles, nil
}

func WalkXmlFile(xmlFile string, l *logger.Logger) ([]string, error) {
	if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
		l.Verbose("Could be an external lib file, skipping: %s", xmlFile)
		return []string{}, nil
	}

	filesReferenced, err := readFilesFromXmlFile(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %v", err)
	}

	var entries []string
	for _, includeFile := range filesReferenced {
		entries = append(entries, includeFile)
		if strings.Contains(includeFile, ".xml") {
			withFixedPathSep := strings.ReplaceAll(includeFile, "/", string(os.PathSeparator))
			withFixedPathSep = strings.ReplaceAll(withFixedPathSep, "\\", string(os.PathSeparator))
			recursiveEntries, err := WalkXmlFile(withFixedPathSep, l)
			if err != nil {
				return nil, fmt.Errorf("error walking XML file: %v", err)
			}
			entries = append(entries, recursiveEntries...)
		}
	}

	return entries, nil
}

func readFilesFromXmlFile(xmlFile string) ([]string, error) {
	contents, err := os.ReadFile(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %v", err)
	}

	files := []string{}
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(line, "<!--") {
			continue
		}
		if strings.Contains(line, "file=") {
			includeFile := strings.Split(line, "file=\"")[1]
			includeFile = strings.Split(includeFile, "\"")[0]
			includeFile = filepath.Join(filepath.Dir(xmlFile), includeFile)
			files = append(files, includeFile)
		}
	}
	return files, nil
}
