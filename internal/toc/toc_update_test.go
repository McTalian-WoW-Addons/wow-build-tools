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

// TestCountAdditions tests counting version additions
func TestCountAdditions(t *testing.T) {
	tests := []struct {
		name        string
		oldVersions []int
		newVersions []int
		expected    int
	}{
		{
			name:        "Single addition",
			oldVersions: []int{110000},
			newVersions: []int{110000, 110007},
			expected:    1,
		},
		{
			name:        "Multiple additions",
			oldVersions: []int{110000},
			newVersions: []int{110000, 110007, 110008},
			expected:    2,
		},
		{
			name:        "No additions",
			oldVersions: []int{110000},
			newVersions: []int{110000},
			expected:    0,
		},
		{
			name:        "All new",
			oldVersions: []int{},
			newVersions: []int{110000, 110007},
			expected:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countAdditions(tt.oldVersions, tt.newVersions)
			if result != tt.expected {
				t.Errorf("Expected %d additions, got %d", tt.expected, result)
			}
		})
	}
}

// TestCountRemovals tests counting version removals
func TestCountRemovals(t *testing.T) {
	tests := []struct {
		name        string
		oldVersions []int
		newVersions []int
		expected    int
	}{
		{
			name:        "Single removal",
			oldVersions: []int{110000, 110001},
			newVersions: []int{110000},
			expected:    1,
		},
		{
			name:        "Multiple removals",
			oldVersions: []int{110000, 110001, 110002},
			newVersions: []int{110000},
			expected:    2,
		},
		{
			name:        "No removals",
			oldVersions: []int{110000},
			newVersions: []int{110000},
			expected:    0,
		},
		{
			name:        "All removed",
			oldVersions: []int{110000, 110001},
			newVersions: []int{},
			expected:    2,
		},
		{
			name:        "Combined add and remove",
			oldVersions: []int{40400},
			newVersions: []int{40401},
			expected:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countRemovals(tt.oldVersions, tt.newVersions)
			if result != tt.expected {
				t.Errorf("Expected %d removals, got %d", tt.expected, result)
			}
		})
	}
}

// TestUpdateResult_JSONMarshaling tests that UpdateResult can be properly marshaled to JSON
func TestUpdateResult_JSONMarshaling(t *testing.T) {
	result := &UpdateResult{
		TotalAdded:   2,
		TotalRemoved: 1,
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
}

// TestUpdateResult_JSONStructure tests that JSON output has expected structure
func TestUpdateResult_JSONStructure(t *testing.T) {
	result := &UpdateResult{
		TotalAdded:   2,
		TotalRemoved: 1,
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

	// Verify values are correct
	if !strings.Contains(jsonStr, "\"total_added\": 2") {
		t.Error("JSON total_added value incorrect")
	}
	if !strings.Contains(jsonStr, "\"total_removed\": 1") {
		t.Error("JSON total_removed value incorrect")
	}
}

// TestUpdateResult_FlavorBreakdown tests that flavor breakdown is properly tracked
func TestUpdateResult_FlavorBreakdown(t *testing.T) {
	result := &UpdateResult{
		TotalAdded:   2,
		TotalRemoved: 1,
		ByFlavor: map[string]FlavorChange{
			"retail": {
				Added:       1,
				Removed:     0,
				OldVersions: []int{120000},
				NewVersions: []int{120000, 120100},
				Products:    []string{"wow", "wow_beta"},
			},
			"mists": {
				Added:       1,
				Removed:     1,
				OldVersions: []int{40400},
				NewVersions: []int{40401},
				Products:    []string{"wow_classic"},
			},
		},
	}

	// Verify structure
	if len(result.ByFlavor) != 2 {
		t.Errorf("Expected 2 flavors, got %d", len(result.ByFlavor))
	}

	// Check retail flavor
	retail := result.ByFlavor["retail"]
	if retail.Added != 1 {
		t.Errorf("Expected retail added=1, got %d", retail.Added)
	}
	if len(retail.NewVersions) != 2 {
		t.Errorf("Expected 2 retail new versions, got %d", len(retail.NewVersions))
	}
	if len(retail.Products) != 2 {
		t.Errorf("Expected 2 retail products, got %d", len(retail.Products))
	}

	// Check mists flavor
	mists := result.ByFlavor["mists"]
	if mists.Added != 1 {
		t.Errorf("Expected mists added=1, got %d", mists.Added)
	}
	if mists.Removed != 1 {
		t.Errorf("Expected mists removed=1, got %d", mists.Removed)
	}

	// Marshal to JSON and verify structure
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify flavor fields exist
	if !strings.Contains(jsonStr, "by_flavor") {
		t.Error("JSON missing by_flavor field")
	}
	if !strings.Contains(jsonStr, "\"retail\"") {
		t.Error("JSON missing retail flavor")
	}
	if !strings.Contains(jsonStr, "\"mists\"") {
		t.Error("JSON missing mists flavor")
	}
	if !strings.Contains(jsonStr, "products") {
		t.Error("JSON missing products field")
	}
}

// TestFlavorChangeAddedRemovedCalculation documents the current calculation behavior
// when old and new versions have overlapping values
func TestFlavorChangeAddedRemovedCalculation(t *testing.T) {
	tests := []struct {
		name            string
		oldVersions     []int
		newVersions     []int
		expectedAdded   int
		expectedRemoved int
		description     string
	}{
		{
			name:            "Version retained and new added",
			oldVersions:     []int{120000},
			newVersions:     []int{120000, 120100},
			expectedAdded:   1,
			expectedRemoved: 0,
			description:     "120000 stays, 120100 is new. Only counts 120100 as added.",
		},
		{
			name:            "Version replaced",
			oldVersions:     []int{40400},
			newVersions:     []int{40401},
			expectedAdded:   1,
			expectedRemoved: 1,
			description:     "40400 removed, 40401 added. Counts as 1 add + 1 remove.",
		},
		{
			name:            "Multiple new versions",
			oldVersions:     []int{120000},
			newVersions:     []int{120000, 120100, 120200},
			expectedAdded:   2,
			expectedRemoved: 0,
			description:     "120000 stays, 120100 and 120200 are new.",
		},
		{
			name:            "Multiple old versions, some removed",
			oldVersions:     []int{120000, 120001, 120002},
			newVersions:     []int{120000, 120100},
			expectedAdded:   1,
			expectedRemoved: 2,
			description:     "120000 stays, 120001/120002 removed, 120100 added.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added := countAdditions(tt.oldVersions, tt.newVersions)
			removed := countRemovals(tt.oldVersions, tt.newVersions)

			if added != tt.expectedAdded {
				t.Errorf("Expected %d added, got %d. %s", tt.expectedAdded, added, tt.description)
			}
			if removed != tt.expectedRemoved {
				t.Errorf("Expected %d removed, got %d. %s", tt.expectedRemoved, removed, tt.description)
			}
		})
	}
}
