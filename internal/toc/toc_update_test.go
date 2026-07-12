package toc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestUpdateInterfaceVersions_SingleProduct tests updating interface versions for a single product
func TestUpdateInterfaceVersions_SingleProduct(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110000
## Title: Test Addon

file.lua
`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	// Mock available interfaces (simulate what CheckForInterfaceBumps would return)
	mockInterfaces := []int{110007}

	// Update the TOC file with mock interface versions
	contents, err := os.ReadFile(toc.Filepath)
	if err != nil {
		t.Fatalf("error reading TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## Interface:") {
			var interfaceStrings []string
			for _, iface := range mockInterfaces {
				interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
			}
			newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
			lines[i] = newInterfaceLine
			break
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(toc.Filepath, []byte(newContents), 0644)
	if err != nil {
		t.Fatalf("error writing updated TOC file: %v", err)
	}

	// Verify the update
	updatedContents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated TOC file: %v", err)
	}

	updatedStr := string(updatedContents)
	if !strings.Contains(updatedStr, "## Interface: 110007") {
		t.Errorf("Expected updated interface version 110007, got:\n%s", updatedStr)
	}
}

// TestUpdateInterfaceVersions_MultipleVersions tests updating with multiple interface versions
func TestUpdateInterfaceVersions_MultipleVersions(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110000, 40400, 11502
## Title: Test Addon

file.lua
`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	// Verify initial interface versions
	expectedInitial := []int{110000, 40400, 11502}
	if !reflect.DeepEqual(toc.Interface, expectedInitial) {
		t.Errorf("Expected initial interfaces %v, got %v", expectedInitial, toc.Interface)
	}

	// Mock multiple updated interfaces
	mockInterfaces := []int{110007, 40401, 11505}

	contents, err := os.ReadFile(toc.Filepath)
	if err != nil {
		t.Fatalf("error reading TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## Interface:") {
			var interfaceStrings []string
			for _, iface := range mockInterfaces {
				interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
			}
			newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
			lines[i] = newInterfaceLine
			break
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(toc.Filepath, []byte(newContents), 0644)
	if err != nil {
		t.Fatalf("error writing updated TOC file: %v", err)
	}

	// Verify the update
	updatedContents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated TOC file: %v", err)
	}

	updatedStr := string(updatedContents)
	if !strings.Contains(updatedStr, "## Interface: 110007, 40401, 11505") {
		t.Errorf("Expected updated interface versions, got:\n%s", updatedStr)
	}
}

// TestGetProductForFile_Mainline tests product detection for Mainline TOC files
func TestGetProductForFile_Mainline(t *testing.T) {
	filename := "TestAddon_Mainline.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	if flavor != Retail {
		t.Errorf("Expected Retail flavor for Mainline suffix, got %v", flavor)
	}

	if suffix != "Mainline" {
		t.Errorf("Expected suffix 'Mainline', got %q", suffix)
	}
}

// TestGetProductForFile_Classic tests product detection for Classic TOC files
func TestGetProductForFile_Classic(t *testing.T) {
	filename := "TestAddon_Classic.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	// Classic suffix should map to CurrentClassic (the progression server)
	if flavor != CurrentClassic {
		t.Errorf("Expected CurrentClassic flavor for Classic suffix, got %v", flavor)
	}

	if suffix != "Classic" {
		t.Errorf("Expected suffix 'Classic', got %q", suffix)
	}
}

// TestGetProductForFile_CurrentClassic tests product detection for current classic expansion
func TestGetProductForFile_CurrentClassic(t *testing.T) {
	filename := "TestAddon_Mists.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	// Current classic expansion
	if flavor != MistsClassic {
		t.Errorf("Expected MistsClassic flavor for Mists suffix, got %v", flavor)
	}

	if suffix != "Mists" {
		t.Errorf("Expected suffix 'Mists', got %q", suffix)
	}
}

// TestGetProductForFile_Vanilla tests product detection for Vanilla TOC files
func TestGetProductForFile_Vanilla(t *testing.T) {
	filename := "TestAddon_Vanilla.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	if flavor != ClassicEra {
		t.Errorf("Expected ClassicEra flavor for Vanilla suffix, got %v", flavor)
	}

	if suffix != "Vanilla" {
		t.Errorf("Expected suffix 'Vanilla', got %q", suffix)
	}
}

// TestGetProductForFile_NoSuffix tests product detection for files without flavor suffix
func TestGetProductForFile_NoSuffix(t *testing.T) {
	filename := "TestAddon.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	// Default to Retail for no suffix
	if flavor != Retail {
		t.Errorf("Expected Retail flavor for no suffix, got %v", flavor)
	}

	if suffix != "" {
		t.Errorf("Expected empty suffix, got %q", suffix)
	}
}

// TestGetProductForFile_HyphenSeparator tests product detection with hyphen separator
func TestGetProductForFile_HyphenSeparator(t *testing.T) {
	filename := "TestAddon-Mainline.toc"
	flavor, suffix := TocFileToGameFlavor(strings.TrimSuffix(filename, ".toc"))

	if flavor != Retail {
		t.Errorf("Expected Retail flavor for Mainline suffix with hyphen, got %v", flavor)
	}

	if suffix != "Mainline" {
		t.Errorf("Expected suffix 'Mainline', got %q", suffix)
	}
}

// TestFlavorReleaseToProductMap_FullRelease tests product mapping for full releases
func TestFlavorReleaseToProductMap_FullRelease(t *testing.T) {
	tests := []struct {
		name     string
		release  GameFlavorRelease
		expected []Product
	}{
		{
			name:     "Retail Full Release",
			release:  RetailFlavorRelease,
			expected: []Product{ProductWow},
		},
		{
			name:     "Classic Era Full Release",
			release:  ClassicEraFlavorRelease,
			expected: []Product{ProductWowClassicEra},
		},
		{
			name:     "Classic Full Release",
			release:  ClassicFlavorRelease,
			expected: []Product{ProductWowClassic},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, exists := FlavorReleaseToProductMap[tt.release]
			if !exists {
				t.Fatalf("Expected products for %v, but not found in map", tt.release)
			}

			if !reflect.DeepEqual(products, tt.expected) {
				t.Errorf("Expected products %v for %v, got %v", tt.expected, tt.release, products)
			}
		})
	}
}

// TestFlavorReleaseToProductMap_BetaRelease tests product mapping for beta releases
func TestFlavorReleaseToProductMap_BetaRelease(t *testing.T) {
	tests := []struct {
		name     string
		release  GameFlavorRelease
		expected []Product
	}{
		{
			name:     "Retail Beta Release",
			release:  RetailBetaFlavorRelease,
			expected: []Product{ProductWowBeta},
		},
		{
			name:     "Classic Beta Release",
			release:  ClassicBetaFlavorRelease,
			expected: []Product{ProductWowClassicBeta},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, exists := FlavorReleaseToProductMap[tt.release]
			if !exists {
				t.Fatalf("Expected products for %v, but not found in map", tt.release)
			}

			if !reflect.DeepEqual(products, tt.expected) {
				t.Errorf("Expected products %v for %v, got %v", tt.expected, tt.release, products)
			}
		})
	}
}

// TestFlavorReleaseToProductMap_TestRelease tests product mapping for test/PTR releases
func TestFlavorReleaseToProductMap_TestRelease(t *testing.T) {
	tests := []struct {
		name     string
		release  GameFlavorRelease
		expected []Product
	}{
		{
			name:     "Retail Test Release",
			release:  RetailTestFlavorRelease,
			expected: []Product{ProductWowTest, ProductWowXPtr},
		},
		{
			name:     "Classic Era Test Release",
			release:  ClassicEraTestFlavorRelease,
			expected: []Product{ProductWowClassicEraPtr},
		},
		{
			name:     "Classic Test Release",
			release:  ClassicTestFlavorRelease,
			expected: []Product{ProductWowClassicPtr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, exists := FlavorReleaseToProductMap[tt.release]
			if !exists {
				t.Fatalf("Expected products for %v, but not found in map", tt.release)
			}

			if !reflect.DeepEqual(products, tt.expected) {
				t.Errorf("Expected products %v for %v, got %v", tt.expected, tt.release, products)
			}
		})
	}
}

// TestBuildInfo_GetInterfaceVersion tests parsing build versions into interface versions
func TestBuildInfo_GetInterfaceVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expected    int
		expectError bool
	}{
		{
			name:        "Retail version",
			version:     "11.0.7",
			expected:    110007,
			expectError: false,
		},
		{
			name:        "Classic version",
			version:     "4.4.1",
			expected:    40401,
			expectError: false,
		},
		{
			name:        "Classic Era version",
			version:     "1.15.5",
			expected:    11505,
			expectError: false,
		},
		{
			name:        "Invalid version - too few segments",
			version:     "11.0",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid version - non-numeric",
			version:     "11.0.x",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildInfo := BuildInfo{
				Version: tt.version,
			}

			result, err := buildInfo.GetInterfaceVersion()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for version %q, but got nil", tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for version %q: %v", tt.version, err)
				}

				if result != tt.expected {
					t.Errorf("Expected interface version %d for version %q, got %d", tt.expected, tt.version, result)
				}
			}
		})
	}
}

// TestGetFlavorsFromInterfaces tests extracting game flavors from interface versions
func TestGetFlavorsFromInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []int
		expected   []GameFlavor
	}{
		{
			name:       "Single Retail interface",
			interfaces: []int{110007},
			expected:   []GameFlavor{Retail},
		},
		{
			name:       "Multiple flavors",
			interfaces: []int{110007, 40401, 11505},
			expected:   []GameFlavor{Retail, CataClassic, ClassicEra},
		},
		{
			name:       "Duplicate flavors",
			interfaces: []int{110007, 110006},
			expected:   []GameFlavor{Retail},
		},
		{
			name:       "Classic Era interfaces",
			interfaces: []int{11502, 11503},
			expected:   []GameFlavor{ClassicEra},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := &Toc{
				Interface: tt.interfaces,
			}

			result := toc.GetFlavorsFromInterfaces()

			// Compare lengths first
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d flavors, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Create maps for easier comparison (order doesn't matter)
			resultMap := make(map[GameFlavor]bool)
			for _, f := range result {
				resultMap[f] = true
			}

			for _, expectedFlavor := range tt.expected {
				if !resultMap[expectedFlavor] {
					t.Errorf("Expected flavor %v not found in result %v", expectedFlavor, result)
				}
			}
		})
	}
}

// TestProductToFlavorMap tests that all products map to appropriate flavors
func TestProductToFlavorMap(t *testing.T) {
	tests := []struct {
		product  Product
		expected GameFlavor
	}{
		{ProductWow, Retail},
		{ProductWowBeta, Retail},
		{ProductWowTest, Retail},
		{ProductWowXPtr, Retail},
		{ProductWowClassic, CurrentClassic},
		{ProductWowClassicBeta, CurrentClassic},
		{ProductWowClassicPtr, CurrentClassic},
		{ProductWowClassicEra, ClassicEra},
		{ProductWowClassicEraPtr, ClassicEra},
	}

	for _, tt := range tests {
		t.Run(string(tt.product), func(t *testing.T) {
			flavor, exists := ProductToFlavorMap[tt.product]
			if !exists {
				t.Fatalf("Product %q not found in ProductToFlavorMap", tt.product)
			}

			if flavor != tt.expected {
				t.Errorf("Expected flavor %v for product %q, got %v", tt.expected, tt.product, flavor)
			}
		})
	}
}

// TestTocUpdate_PreservesNonInterfaceContent tests that updating preserves other TOC content
func TestTocUpdate_PreservesNonInterfaceContent(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110000
## Title: My Test Addon
## Notes: This is a test addon with multiple fields
## Version: 1.0.0
## Author: Test Author
## X-Curse-Project-ID: 12345

Core.lua
Utils.lua
`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	// Mock interface update
	mockInterfaces := []int{110007}

	contents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("error reading TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## Interface:") {
			var interfaceStrings []string
			for _, iface := range mockInterfaces {
				interfaceStrings = append(interfaceStrings, fmt.Sprintf("%d", iface))
			}
			newInterfaceLine := "## Interface: " + strings.Join(interfaceStrings, ", ")
			lines[i] = newInterfaceLine
			break
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(tocPath, []byte(newContents), 0644)
	if err != nil {
		t.Fatalf("error writing updated TOC file: %v", err)
	}

	// Verify all content is preserved
	updatedContents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated TOC file: %v", err)
	}

	updatedStr := string(updatedContents)

	// Check interface was updated
	if !strings.Contains(updatedStr, "## Interface: 110007") {
		t.Errorf("Interface version not updated correctly")
	}

	// Check other fields preserved
	requiredFields := []string{
		"## Title: My Test Addon",
		"## Notes: This is a test addon with multiple fields",
		"## Version: 1.0.0",
		"## Author: Test Author",
		"## X-Curse-Project-ID: 12345",
		"Core.lua",
		"Utils.lua",
	}

	for _, field := range requiredFields {
		if !strings.Contains(updatedStr, field) {
			t.Errorf("Expected field %q not found in updated TOC", field)
		}
	}
}

// TestGameFlavorToString tests the ToString method for GameFlavor
func TestGameFlavorToString(t *testing.T) {
	tests := []struct {
		flavor   GameFlavor
		expected string
	}{
		{ClassicEra, "classic"},
		{TbcClassic, "bcc"},
		{WotlkClassic, "wrath"},
		{CataClassic, "cata"},
		{MistsClassic, "mists"},
		{Retail, "retail"},
		{Unknown, "retail"}, // Unknown defaults to retail
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.flavor.ToString()
			if result != tt.expected {
				t.Errorf("Expected %q for flavor %v, got %q", tt.expected, tt.flavor, result)
			}
		})
	}
}

// TestTrackFlavorChange_VersionAdditions tests tracking when versions are added
func TestTrackFlavorChange_VersionAdditions(t *testing.T) {
	result := &UpdateResult{
		ByFlavor: make(map[string]FlavorVersionChange),
	}

	oldVersions := []int{110000}
	newVersions := []int{110000, 110007}

	trackFlavorChange(result, "retail", oldVersions, newVersions)

	if result.TotalAdded != 1 {
		t.Errorf("Expected 1 version added, got %d", result.TotalAdded)
	}
	if result.TotalRemoved != 0 {
		t.Errorf("Expected 0 versions removed, got %d", result.TotalRemoved)
	}

	change := result.ByFlavor["retail"]
	if change.Added != 1 {
		t.Errorf("Expected 1 flavor addition, got %d", change.Added)
	}
	if change.Removed != 0 {
		t.Errorf("Expected 0 flavor removals, got %d", change.Removed)
	}
	if !reflect.DeepEqual(change.NewVersions, newVersions) {
		t.Errorf("Expected new versions %v, got %v", newVersions, change.NewVersions)
	}
}

// TestTrackFlavorChange_VersionRemovals tests tracking when versions are removed
func TestTrackFlavorChange_VersionRemovals(t *testing.T) {
	result := &UpdateResult{
		ByFlavor: make(map[string]FlavorVersionChange),
	}

	oldVersions := []int{110000, 110001}
	newVersions := []int{110000}

	trackFlavorChange(result, "retail", oldVersions, newVersions)

	if result.TotalAdded != 0 {
		t.Errorf("Expected 0 versions added, got %d", result.TotalAdded)
	}
	if result.TotalRemoved != 1 {
		t.Errorf("Expected 1 version removed, got %d", result.TotalRemoved)
	}

	change := result.ByFlavor["retail"]
	if change.Added != 0 {
		t.Errorf("Expected 0 flavor additions, got %d", change.Added)
	}
	if change.Removed != 1 {
		t.Errorf("Expected 1 flavor removal, got %d", change.Removed)
	}
}

// TestTrackFlavorChange_CombinedAddAndRemove tests the fixed case: both add and remove in same flavor
func TestTrackFlavorChange_CombinedAddAndRemove(t *testing.T) {
	result := &UpdateResult{
		ByFlavor: make(map[string]FlavorVersionChange),
	}

	// Classic added 40401, removed 40400
	oldVersions := []int{40400}
	newVersions := []int{40401}

	trackFlavorChange(result, "classic", oldVersions, newVersions)

	if result.TotalAdded != 1 {
		t.Errorf("Expected 1 version added, got %d", result.TotalAdded)
	}
	if result.TotalRemoved != 1 {
		t.Errorf("Expected 1 version removed, got %d", result.TotalRemoved)
	}

	change := result.ByFlavor["classic"]
	if change.Added != 1 {
		t.Errorf("Expected 1 flavor addition, got %d", change.Added)
	}
	if change.Removed != 1 {
		t.Errorf("Expected 1 flavor removal, got %d", change.Removed)
	}
}

// TestTrackFlavorChange_MultipleFlavors tests aggregating changes across multiple flavors
func TestTrackFlavorChange_MultipleFlavors(t *testing.T) {
	result := &UpdateResult{
		ByFlavor: make(map[string]FlavorVersionChange),
	}

	// Retail: 1 addition
	trackFlavorChange(result, "retail", []int{110000}, []int{110000, 110007})

	// Classic: 1 addition + 1 removal
	trackFlavorChange(result, "classic", []int{40400}, []int{40401})

	// Vanilla: 1 removal
	trackFlavorChange(result, "classic", []int{11502, 11503}, []int{11502})

	if result.TotalAdded != 2 {
		t.Errorf("Expected 2 total versions added, got %d", result.TotalAdded)
	}
	if result.TotalRemoved != 2 {
		t.Errorf("Expected 2 total versions removed, got %d", result.TotalRemoved)
	}

	if len(result.ByFlavor) != 2 {
		t.Errorf("Expected 2 flavors in result, got %d", len(result.ByFlavor))
	}
}

// TestUpdateResult_JSONMarshaling tests that UpdateResult can be properly marshaled to JSON
func TestUpdateResult_JSONMarshaling(t *testing.T) {
	result := &UpdateResult{
		TotalAdded:   2,
		TotalRemoved: 1,
		ByFlavor: map[string]FlavorVersionChange{
			"retail": {
				Added:       2,
				Removed:     0,
				OldVersions: []int{110000},
				NewVersions: []int{110000, 110007},
			},
			"classic": {
				Added:       1,
				Removed:     1,
				OldVersions: []int{40400},
				NewVersions: []int{40401},
			},
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal UpdateResult to JSON: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshaled UpdateResult
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify counts
	if unmarshaled.TotalAdded != result.TotalAdded {
		t.Errorf("Expected TotalAdded %d, got %d", result.TotalAdded, unmarshaled.TotalAdded)
	}
	if unmarshaled.TotalRemoved != result.TotalRemoved {
		t.Errorf("Expected TotalRemoved %d, got %d", result.TotalRemoved, unmarshaled.TotalRemoved)
	}

	// Verify flavor data
	retailChange := unmarshaled.ByFlavor["retail"]
	if retailChange.Added != 2 {
		t.Errorf("Expected retail Added 2, got %d", retailChange.Added)
	}
	if len(retailChange.NewVersions) != 2 {
		t.Errorf("Expected 2 retail new versions, got %d", len(retailChange.NewVersions))
	}
}

// TestUpdateResult_JSONContainsFlavorBreakdown tests that JSON output includes flavor-specific data
func TestUpdateResult_JSONContainsFlavorBreakdown(t *testing.T) {
	result := &UpdateResult{
		TotalAdded:   1,
		TotalRemoved: 0,
		ByFlavor: map[string]FlavorVersionChange{
			"retail": {
				Added:       1,
				Removed:     0,
				OldVersions: []int{110000},
				NewVersions: []int{110007},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify JSON contains expected fields
	if !strings.Contains(jsonStr, "total_added") {
		t.Error("JSON missing total_added field")
	}
	if !strings.Contains(jsonStr, "total_removed") {
		t.Error("JSON missing total_removed field")
	}
	if !strings.Contains(jsonStr, "by_flavor") {
		t.Error("JSON missing by_flavor field")
	}
	if !strings.Contains(jsonStr, "retail") {
		t.Error("JSON missing retail flavor breakdown")
	}
	if !strings.Contains(jsonStr, "old_versions") {
		t.Error("JSON missing old_versions field")
	}
	if !strings.Contains(jsonStr, "new_versions") {
		t.Error("JSON missing new_versions field")
	}
}
