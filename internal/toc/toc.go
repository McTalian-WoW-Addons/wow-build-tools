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

type Toc struct {
	Filepath              string
	Interface             []int
	Title                 string
	Notes                 string
	Version               string
	Files                 []string
	CurseId               string
	WowiId                string
	WagoId                string
	Flavor                GameFlavor
	tocSpecificInterfaces map[GameFlavor][]int
}

// UpdateResult contains metadata about version changes
type UpdateResult struct {
	TotalAdded   int `json:"total_added"`
	TotalRemoved int `json:"total_removed"`
}

var l = logger.DefaultLogger

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
		t.tocSpecificInterfaces[flavor] = append(t.tocSpecificInterfaces[flavor], interfaceVersion)
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

func (t *Toc) getProductsToCheck(flavorReleaseInfo FlavorReleaseInfo) (productsToCheck []Product, err error) {
	_, err = GetLatestBuildInfo()
	if err != nil {
		return
	}
	var releaseTypes = []GameReleaseType{LiveRelease}
	if flavorReleaseInfo.IsBeta {
		releaseTypes = append(releaseTypes, BetaRelease)
	}
	if flavorReleaseInfo.IsTest {
		releaseTypes = append(releaseTypes, TestRelease)
	}

	for flavor := range t.tocSpecificInterfaces {
		for _, releaseType := range releaseTypes {
			flavorRelease := GameFlavorRelease{
				Flavor:      flavor,
				ReleaseType: releaseType,
			}
			products, exists := FlavorReleaseToProductMap[flavorRelease]
			if exists {
				productsToCheck = append(productsToCheck, products...)
			} else {
				l.Warn("No products found for flavor release: %s", flavorRelease.ToString())
			}
		}
	}

	return
}

func (t *Toc) CheckForInterfaceBumps(flavorReleaseInfo FlavorReleaseInfo) (availableInterfaces map[Product]int, err error) {
	productsToCheck, err := t.getProductsToCheck(flavorReleaseInfo)
	if err != nil {
		return
	}

	availableInterfaces = make(map[Product]int)

	for _, product := range productsToCheck {
		buildInfo, exists := (*cacheLatestBuilds)[product]
		if !exists {
			continue
		}
		interfaceVersion, err := buildInfo.GetInterfaceVersion()
		if err != nil {
			return nil, fmt.Errorf("error parsing Interface version for product %s: %v", product, err)
		}
		if product.IsBeta() || product.IsTest() {
			liveProduct := product.GetLive()
			liveInterface, liveExists := availableInterfaces[liveProduct]
			if !liveExists {
				liveBuildInfo, liveCacheExists := (*cacheLatestBuilds)[liveProduct]
				if liveCacheExists {
					liveInterface, err = liveBuildInfo.GetInterfaceVersion()
					if err != nil {
						return nil, fmt.Errorf("error parsing Interface version for live product %s: %v", liveProduct, err)
					}
				}
			}

			if liveInterface > interfaceVersion {
				// Skip older version
				continue
			}
		}

		availableInterfaces[product] = interfaceVersion
	}

	return
}

func (t *Toc) UpdateInterfaceVersions(flavorReleaseInfo FlavorReleaseInfo) (*UpdateResult, error) {
	result := &UpdateResult{}

	availableInterfaces, err := t.CheckForInterfaceBumps(flavorReleaseInfo)
	if err != nil {
		return nil, fmt.Errorf("error checking for interface bumps: %v", err)
	}

	// Update the toc file with the new interface versions
	contents, err := os.ReadFile(t.Filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")

	// Store old interface versions before update
	oldInterfaces := make(map[string][]int)
	for flavor := range t.tocSpecificInterfaces {
		oldInterfaces[flavor.ToString()] = append([]int{}, t.tocSpecificInterfaces[flavor]...)
	}

	// Detect if file uses single-line or multi-line format
	hasSuffixedLines := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "## Interface-") {
			hasSuffixedLines = true
			break
		}
	}

	if hasSuffixedLines {
		// Multi-line format: update each suffixed line with the appropriate flavor's interface
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "## Interface:") && !strings.Contains(trimmedLine, "## Interface-") {
				// Update the main Interface line with all versions, sorted
				var interfaces []int
				for _, iface := range availableInterfaces {
					interfaces = append(interfaces, iface)
				}
				slices.Sort(interfaces)
				var interfaceStrings []string
				for _, iface := range interfaces {
					interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
				}
				newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
				lines[i] = newInterfaceLine
			} else if strings.HasPrefix(trimmedLine, "## Interface-") {
				// Update suffixed Interface lines only for currently active flavors
				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) == 2 {
					// Extract the suffix (e.g., "Vanilla", "Classic", "Mainline", "Mists")
					suffix := strings.TrimPrefix(parts[0], "## Interface-")
					// TocFileToGameFlavor expects a filename-like string with dash/underscore
					// So we prepend a dummy name to make it work
					flavor, _ := TocFileToGameFlavor("Addon-" + suffix)

					// Normalize MistsClassic to CurrentClassic for comparison
					if flavor == MistsClassic {
						flavor = CurrentClassic
					}

					var flavorReleases = []GameFlavorRelease{
						{
							Flavor:      flavor,
							ReleaseType: LiveRelease,
						},
					}
					if flavorReleaseInfo.IsBeta {
						flavorReleases = append(flavorReleases, GameFlavorRelease{
							Flavor:      flavor,
							ReleaseType: BetaRelease,
						})
					}
					if flavorReleaseInfo.IsTest {
						flavorReleases = append(flavorReleases, GameFlavorRelease{
							Flavor:      flavor,
							ReleaseType: TestRelease,
						})
					}

					newIfaces := []int{}
					products := []Product{}
					for _, flavorRelease := range flavorReleases {
						flavorProducts, exists := FlavorReleaseToProductMap[flavorRelease]
						if exists {
							products = append(products, flavorProducts...)
						}
					}

					productMap := make(map[Product]bool)
					for _, product := range products {
						productMap[product] = true
					}

					for product := range productMap {
						// Only update if we have an interface version for this flavor
						if iface, exists := availableInterfaces[product]; exists {
							newIfaces = append(newIfaces, iface)
						}
					}

					// Only update the line if we found interface versions for this flavor
					if len(newIfaces) > 0 {
						iFaceMap := make(map[int]bool)
						for _, iface := range newIfaces {
							iFaceMap[iface] = true
						}

						// Remove duplicates, convert to strings
						uniqueIfaces := []string{}
						for iface := range iFaceMap {
							uniqueIfaces = append(uniqueIfaces, strconv.Itoa(iface))
						}

						slices.Sort(uniqueIfaces)
						newIface := strings.Join(uniqueIfaces, ", ")
						lines[i] = fmt.Sprintf("## Interface-%s: %s", suffix, newIface)

						// Track changes
						flavorStr := flavor.ToString()
						oldVer := oldInterfaces[flavorStr]
						newVer := []int{}
						for _, v := range uniqueIfaces {
							val, _ := strconv.Atoi(v)
							newVer = append(newVer, val)
						}
						result.TotalAdded += countAdditions(oldVer, newVer)
						result.TotalRemoved += countRemovals(oldVer, newVer)
					}
					// If no new interfaces were found for this flavor, leave the line unchanged
				}
			}
		}
	} else {
		// Single-line format: update only the main Interface line
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "## Interface:") {
				// Sort interfaces for consistent output
				var interfaces []int
				for _, iface := range availableInterfaces {
					interfaces = append(interfaces, iface)
				}

				var iFaceMap = make(map[int]bool)
				for _, iface := range interfaces {
					iFaceMap[iface] = true
				}

				// Remove duplicates
				interfaces = []int{}
				for iface := range iFaceMap {
					interfaces = append(interfaces, iface)
				}

				slices.Sort(interfaces)
				var interfaceStrings []string
				for _, iface := range interfaces {
					interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
				}
				newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
				lines[i] = newInterfaceLine

				// Track changes
				oldVer := t.Interface
				result.TotalAdded += countAdditions(oldVer, interfaces)
				result.TotalRemoved += countRemovals(oldVer, interfaces)

				break
			}
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(t.Filepath, []byte(newContents), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing updated TOC file: %v", err)
	}

	return result, nil
}

// countAdditions counts how many versions were added
func countAdditions(oldVersions, newVersions []int) int {
	oldSet := make(map[int]bool)
	for _, v := range oldVersions {
		oldSet[v] = true
	}

	added := 0
	for _, v := range newVersions {
		if !oldSet[v] {
			added++
		}
	}
	return added
}

// countRemovals counts how many versions were removed
func countRemovals(oldVersions, newVersions []int) int {
	newSet := make(map[int]bool)
	for _, v := range newVersions {
		newSet[v] = true
	}

	removed := 0
	for _, v := range oldVersions {
		if !newSet[v] {
			removed++
		}
	}
	return removed
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
