package toc

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

func TestNewToc_ValidFile(t *testing.T) {
	// Create a temporary TOC file for testing
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 111000, 40400
## Title: Test Addon
## Notes: A test addon for unit testing
## Version: 1.0.0
## X-Curse-Project-ID: 12345
## X-WoWI-ID: 67890
## X-Wago-ID: test-addon

Core.lua
Utils.lua
GUI/MainFrame.xml`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	// Test NewToc function
	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	// Verify TOC fields
	if toc.Filepath != tocPath {
		t.Errorf("Expected Filepath %q, got %q", tocPath, toc.Filepath)
	}

	expectedInterface := []int{111000, 40400}
	if !reflect.DeepEqual(toc.Interface, expectedInterface) {
		t.Errorf("Expected Interface %v, got %v", expectedInterface, toc.Interface)
	}

	if toc.Title != "Test Addon" {
		t.Errorf("Expected Title %q, got %q", "Test Addon", toc.Title)
	}

	if toc.Notes != "A test addon for unit testing" {
		t.Errorf("Expected Notes %q, got %q", "A test addon for unit testing", toc.Notes)
	}

	if toc.Version != "1.0.0" {
		t.Errorf("Expected Version %q, got %q", "1.0.0", toc.Version)
	}

	if toc.CurseId != "12345" {
		t.Errorf("Expected CurseId %q, got %q", "12345", toc.CurseId)
	}

	if toc.WowiId != "67890" {
		t.Errorf("Expected WowiId %q, got %q", "67890", toc.WowiId)
	}

	if toc.WagoId != "test-addon" {
		t.Errorf("Expected WagoId %q, got %q", "test-addon", toc.WagoId)
	}

	expectedFiles := []string{"Core.lua", "Utils.lua", "GUI/MainFrame.xml"}
	if !reflect.DeepEqual(toc.Files, expectedFiles) {
		t.Errorf("Expected Files %v, got %v", expectedFiles, toc.Files)
	}

	if toc.Flavor != Retail {
		t.Errorf("Expected Flavor %v, got %v", Retail, toc.Flavor)
	}
}

func TestNewToc_NonExistentFile(t *testing.T) {
	_, err := NewToc("/nonexistent/path/addon.toc")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "error reading TOC file") {
		t.Errorf("Expected error message to contain 'error reading TOC file', got %q", err.Error())
	}
}

func TestNewToc_InvalidInterface(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestInvalid.toc")

	tocContent := `## Interface: invalid
## Title: Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	_, err = NewToc(tocPath)
	if err == nil {
		t.Error("Expected error for invalid interface version, got nil")
	}

	if !strings.Contains(err.Error(), "error parsing Interface version") {
		t.Errorf("Expected error message to contain 'error parsing Interface version', got %q", err.Error())
	}
}

func TestNewToc_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "Empty.toc")

	err := os.WriteFile(tocPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed for empty file: %v", err)
	}

	if len(toc.Interface) != 0 {
		t.Errorf("Expected empty Interface slice, got %v", toc.Interface)
	}

	if len(toc.Files) != 0 {
		t.Errorf("Expected empty Files slice, got %v", toc.Files)
	}
}

func TestToc_addGameVersionsFromToc(t *testing.T) {
	// Reset gameVersions and gameInterfaces before test
	gameVersions = make(map[GameFlavor][]string)
	gameInterfaces = make(map[GameFlavor][]int)

	toc := &Toc{
		Interface:             []int{111000, 50500, 11502}, // Retail, Current Classic, Classic Era
		tocSpecificInterfaces: make(map[GameFlavor][]int),
	}

	result := toc.addGameVersionsFromToc()

	// Check that gameVersions were populated
	if len(result) == 0 {
		t.Error("Expected gameVersions to be populated, got empty map")
	}

	// Test interface version parsing
	// 111000: major=11, minor=10, patch=0 -> Retail
	// 40400: major=4, minor=4, patch=0 -> Classic Era
	// 11502: major=1, minor=15, patch=2 -> Classic Era

	if len(gameVersions[Retail]) == 0 {
		t.Error("Expected Retail game versions to be populated")
	}

	if len(gameVersions[ClassicEra]) == 0 {
		t.Error("Expected Classic Era game versions to be populated")
	}
}

func TestToc_GetTocEntriesTree(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	coreFile := filepath.Join(tempDir, "Core.lua")
	utilsFile := filepath.Join(tempDir, "Utils.lua")
	ignoredFile := filepath.Join(tempDir, "Ignored.lua")

	err := os.WriteFile(coreFile, []byte("-- Core file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create Core.lua: %v", err)
	}

	err = os.WriteFile(utilsFile, []byte("-- Utils file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create Utils.lua: %v", err)
	}

	err = os.WriteFile(ignoredFile, []byte("-- Ignored file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create Ignored.lua: %v", err)
	}

	toc := &Toc{
		Files: []string{"Core.lua", "Utils.lua", "Ignored.lua"},
	}

	logger := logger.GetSubLog("test")

	// Test without ignored files
	tree, err := toc.GetTocEntriesTree(tempDir, []string{}, logger)
	if err != nil {
		t.Fatalf("GetTocEntriesTree failed: %v", err)
	}

	if len(tree.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(tree.Entries))
	}

	expectedPaths := []string{
		filepath.Join(tempDir, "Core.lua"),
		filepath.Join(tempDir, "Utils.lua"),
		filepath.Join(tempDir, "Ignored.lua"),
	}

	for i, entry := range tree.Entries {
		if entry.Filepath != expectedPaths[i] {
			t.Errorf("Expected entry %d to have filepath %q, got %q", i, expectedPaths[i], entry.Filepath)
		}
	}

	// Test with ignored files
	tree, err = toc.GetTocEntriesTree(tempDir, []string{"Ignored.lua"}, logger)
	if err != nil {
		t.Fatalf("GetTocEntriesTree failed with ignored files: %v", err)
	}

	if len(tree.Entries) != 2 {
		t.Errorf("Expected 2 entries after ignoring one file, got %d", len(tree.Entries))
	}

	// Check that ignored file is not in the tree
	for _, entry := range tree.Entries {
		if strings.HasSuffix(entry.Filepath, "Ignored.lua") {
			t.Error("Ignored.lua should not be in the TOC tree")
		}
	}
}

func TestToc_GetTocEntriesTree_WithXMLFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple XML file (actual XML parsing is tested elsewhere)
	xmlFile := filepath.Join(tempDir, "embed.xml")
	xmlContent := `<Ui xmlns="http://www.blizzard.com/wow/ui/">
</Ui>`

	err := os.WriteFile(xmlFile, []byte(xmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create XML file: %v", err)
	}

	toc := &Toc{
		Files: []string{"embed.xml"},
	}

	logger := logger.GetSubLog("test")

	tree, err := toc.GetTocEntriesTree(tempDir, []string{}, logger)
	if err != nil {
		t.Fatalf("GetTocEntriesTree failed with XML: %v", err)
	}

	if len(tree.Entries) != 1 {
		t.Errorf("Expected 1 top-level entry (embed.xml), got %d", len(tree.Entries))
	}

	// Verify that XML entry was created with correct path
	xmlEntry := tree.Entries[0]
	expectedPath := filepath.Join(tempDir, "embed.xml")
	if xmlEntry.Filepath != expectedPath {
		t.Errorf("Expected XML entry path %q, got %q", expectedPath, xmlEntry.Filepath)
	}
}

func TestNewToc_ClassicFlavor(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon-Classic.toc")

	tocContent := `## Interface: 50000
## Title: Classic Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	if toc.Flavor != CurrentClassic {
		t.Errorf("Expected Flavor %v for Classic TOC (CurrentClassic), got %v", CurrentClassic, toc.Flavor)
	}
}

func TestNewToc_MultipleInterfaceVersions(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "MultiVersion.toc")

	tocContent := `## Interface: 111000, 40400, 30403
## Title: Multi Version Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	expectedInterfaces := []int{111000, 40400, 30403}
	if !reflect.DeepEqual(toc.Interface, expectedInterfaces) {
		t.Errorf("Expected Interface versions %v, got %v", expectedInterfaces, toc.Interface)
	}
}

func TestNewToc_CommentsAndEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "CommentsTest.toc")

	tocContent := `## Interface: 111000
## Title: Test Addon
# This is a comment and should be ignored
## Notes: Test notes

# Another comment
Core.lua

Utils.lua
# Final comment`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	// Should only have the two non-comment files
	expectedFiles := []string{"Core.lua", "Utils.lua"}
	if !reflect.DeepEqual(toc.Files, expectedFiles) {
		t.Errorf("Expected Files %v, got %v", expectedFiles, toc.Files)
	}

	if toc.Title != "Test Addon" {
		t.Errorf("Expected Title %q, got %q", "Test Addon", toc.Title)
	}

	if toc.Notes != "Test notes" {
		t.Errorf("Expected Notes %q, got %q", "Test notes", toc.Notes)
	}
}

func TestToc_CheckForInterfaceBumpsNormal(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110000, 50500, 11000
## Title: Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	flavorReleaseInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: false,
	}

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	if err != nil {
		t.Fatalf("CheckForInterfaceBumps failed: %v", err)
	}

	if len(availableInterfaces) == 0 {
		t.Error("Expected at least one available interface version")
	}

	if len(availableInterfaces) != 3 {
		t.Errorf("Expected exactly three available interface versions, got %d", len(availableInterfaces))
	}

	// Check that newer versions are available (they should be greater than our test versions)
	if retailIface, exists := availableInterfaces[Retail]; !exists || retailIface <= 110000 {
		t.Errorf("Expected available interface version for retail to be greater than 110000, got %d", retailIface)
	}
	if classicIface, exists := availableInterfaces[CurrentClassic]; !exists || classicIface <= 50500 {
		t.Errorf("Expected available interface version for current classic to be greater than 50500, got %d", classicIface)
	}
	if eraIface, exists := availableInterfaces[ClassicEra]; !exists || eraIface <= 11000 {
		t.Errorf("Expected available interface version for classic era to be greater than 11000, got %d", eraIface)
	}
}

func TestToc_CheckForInterfaceBumpsPtr(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110205, 50010, 11010
## Title: Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	flavorReleaseInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: true,
	}

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	if err != nil {
		t.Fatalf("CheckForInterfaceBumps failed: %v", err)
	}

	if len(availableInterfaces) == 0 {
		t.Error("Expected at least one available interface version")
	}

	// When IsTest is true, we should get the latest versions from test/PTR products
	// Check that we have versions for all three flavors
	if len(availableInterfaces) < 3 {
		t.Errorf("Expected at least 3 flavors when IsTest=true, got %d", len(availableInterfaces))
	}

	// Verify that we have entries for the expected flavors
	if _, exists := availableInterfaces[Retail]; !exists {
		t.Error("Expected Retail flavor in results")
	}
	if _, exists := availableInterfaces[CurrentClassic]; !exists {
		t.Error("Expected CurrentClassic flavor in results")
	}
	if _, exists := availableInterfaces[ClassicEra]; !exists {
		t.Error("Expected ClassicEra flavor in results")
	}
}

func TestToc_CheckForInterfaceBumpsBeta(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110205, 50010, 11010
## Title: Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	flavorReleaseInfo := FlavorReleaseInfo{
		IsBeta: true,
		IsTest: false,
	}

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	if err != nil {
		t.Fatalf("CheckForInterfaceBumps failed: %v", err)
	}

	if len(availableInterfaces) == 0 {
		t.Error("Expected at least one available interface version")
	}

	// When IsBeta is true, we should get the latest versions from beta products
	// Check that we have versions for all three flavors
	if len(availableInterfaces) < 3 {
		t.Errorf("Expected at least 3 flavors when IsBeta=true, got %d", len(availableInterfaces))
	}

	// Verify that we have entries for the expected flavors
	if _, exists := availableInterfaces[Retail]; !exists {
		t.Error("Expected Retail flavor in results")
	}
	if _, exists := availableInterfaces[CurrentClassic]; !exists {
		t.Error("Expected CurrentClassic flavor in results")
	}
	if _, exists := availableInterfaces[ClassicEra]; !exists {
		t.Error("Expected ClassicEra flavor in results")
	}
}

func TestToc_CheckForInterfaceBumpsBetaAndPtr(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "TestAddon.toc")

	tocContent := `## Interface: 110205, 50010, 11010
## Title: Test Addon

Core.lua`

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("NewToc failed: %v", err)
	}

	flavorReleaseInfo := FlavorReleaseInfo{
		IsBeta: true,
		IsTest: true,
	}

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	if err != nil {
		t.Fatalf("CheckForInterfaceBumps failed: %v", err)
	}

	if len(availableInterfaces) == 0 {
		t.Error("Expected at least one available interface version")
	}

	// When both IsBeta and IsTest are true, we get the highest version from all release types
	// Still expect one version per flavor (the highest found across all release types)
	if len(availableInterfaces) < 3 {
		t.Errorf("Expected at least 3 flavors when IsBeta=true and IsTest=true, got %d", len(availableInterfaces))
	}

	// Verify that we have entries for the expected flavors
	if _, exists := availableInterfaces[Retail]; !exists {
		t.Error("Expected Retail flavor in results")
	}
	if _, exists := availableInterfaces[CurrentClassic]; !exists {
		t.Error("Expected CurrentClassic flavor in results")
	}
	if _, exists := availableInterfaces[ClassicEra]; !exists {
		t.Error("Expected ClassicEra flavor in results")
	}
}
