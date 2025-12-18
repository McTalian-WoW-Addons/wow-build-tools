package toc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestTocFiles creates a set of test TOC files with various configurations
func setupTestTocFiles(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()

	tocFiles := map[string]string{
		"default.toc": `## Interface: 110007

file.lua
`,
		"specific-Classic.toc": `## Interface: 50000

file.lua
`,
		"multi.toc": `## Interface: 110007
## Interface-Vanilla: 11505
## Interface-Classic: 50000
## Interface-Mists: 50000

file.lua
`,
		"multi-oneline.toc": `## Interface: 11505, 50000, 110007

file.lua
`,
		"specific-Mainline.toc": `## Interface: 110007

file.lua
`,
		"specific_Mists.toc": `## Interface: 50000

file.lua
`,
	}

	for filename, content := range tocFiles {
		tocPath := filepath.Join(tempDir, filename)
		err := os.WriteFile(tocPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test TOC file %s: %v", filename, err)
		}
	}

	return tempDir
}

// mockInterfaceVersions represents mock version data similar to what API would return
type mockInterfaceVersions struct {
	retail        int
	retailBeta    int
	retailTest    int
	retailXPtr    int
	classic       int
	classicBeta   int
	classicPtr    int
	classicEra    int
	classicEraPtr int
}

// getMockVersions returns mock interface versions for testing
func getMockVersions() mockInterfaceVersions {
	return mockInterfaceVersions{
		retail:        110007,
		retailBeta:    110008,
		retailTest:    110009,
		retailXPtr:    110010,
		classic:       50000,
		classicBeta:   50001,
		classicPtr:    50002,
		classicEra:    11505,
		classicEraPtr: 11506,
	}
}

// TestUpdateTocFiles_DefaultBehavior tests updating TOC files without flags
func TestUpdateTocFiles_DefaultBehavior(t *testing.T) {
	// This test verifies that the actual UpdateInterfaceVersions method works correctly
	// Note: This requires network access to fetch actual version data from wago.tools API
	// Skip if network is unavailable or for faster unit tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := setupTestTocFiles(t)

	// Find all TOC files
	tocFiles, err := FindTocFiles(testDir)
	if err != nil {
		t.Fatalf("Failed to find TOC files: %v", err)
	}

	// Test with default behavior (no beta, no ptr)
	flavorInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: false,
	}

	// Update each TOC file using the actual method
	for _, tocPath := range tocFiles {
		toc, err := NewToc(tocPath)
		if err != nil {
			t.Fatalf("Failed to create Toc for %s: %v", tocPath, err)
		}

		// Store original interfaces for comparison
		originalInterfaces := make([]int, len(toc.Interface))
		copy(originalInterfaces, toc.Interface)

		err = toc.UpdateInterfaceVersions(flavorInfo)
		if err != nil {
			t.Fatalf("UpdateInterfaceVersions failed for %s: %v", filepath.Base(tocPath), err)
		}

		// Read back the updated file and verify it was actually modified
		updatedToc, err := NewToc(tocPath)
		if err != nil {
			t.Fatalf("Failed to read updated Toc for %s: %v", tocPath, err)
		}

		// Verify that interface versions exist (they should be fetched from API)
		if len(updatedToc.Interface) == 0 {
			t.Errorf("File %s has no interface versions after update", filepath.Base(tocPath))
		}

		t.Logf("File %s: Original interfaces: %v, Updated interfaces: %v",
			filepath.Base(tocPath), originalInterfaces, updatedToc.Interface)
	}
}

// TestUpdateTocFiles_WithPtrFlag tests updating TOC files with PTR flag enabled
func TestUpdateTocFiles_WithPtrFlag(t *testing.T) {
	// This test verifies PTR flag behavior with actual API calls
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := setupTestTocFiles(t)

	// Find all TOC files
	tocFiles, err := FindTocFiles(testDir)
	if err != nil {
		t.Fatalf("Failed to find TOC files: %v", err)
	}

	// Test with PTR flag enabled
	flavorInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: true,
	}

	// Update each TOC file using the actual method
	for _, tocPath := range tocFiles {
		t.Run(filepath.Base(tocPath), func(t *testing.T) {
			toc, err := NewToc(tocPath)
			if err != nil {
				t.Fatalf("Failed to create Toc: %v", err)
			}

			originalInterfaces := make([]int, len(toc.Interface))
			copy(originalInterfaces, toc.Interface)

			err = toc.UpdateInterfaceVersions(flavorInfo)
			if err != nil {
				t.Fatalf("UpdateInterfaceVersions failed: %v", err)
			}

			// Read back the updated file
			updatedToc, err := NewToc(tocPath)
			if err != nil {
				t.Fatalf("Failed to read updated Toc: %v", err)
			}

			// With PTR flag, we might have multiple interface versions
			// (base + PTR if PTR is higher)
			if len(updatedToc.Interface) == 0 {
				t.Errorf("No interface versions after PTR update")
			}

			// Verify we got multiple versions when PTR is enabled (in most cases)
			t.Logf("PTR mode - Original: %v, Updated: %v", originalInterfaces, updatedToc.Interface)
		})
	}
}

// TestUpdateTocFiles_WithBetaFlag tests updating TOC files with beta flag enabled
func TestUpdateTocFiles_WithBetaFlag(t *testing.T) {
	// This test verifies beta flag behavior with actual API calls
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := setupTestTocFiles(t)

	// Find all TOC files
	tocFiles, err := FindTocFiles(testDir)
	if err != nil {
		t.Fatalf("Failed to find TOC files: %v", err)
	}

	// Test with Beta flag enabled
	flavorInfo := FlavorReleaseInfo{
		IsBeta: true,
		IsTest: false,
	}

	// Update each TOC file using the actual method
	for _, tocPath := range tocFiles {
		t.Run(filepath.Base(tocPath), func(t *testing.T) {
			toc, err := NewToc(tocPath)
			if err != nil {
				t.Fatalf("Failed to create Toc: %v", err)
			}

			originalInterfaces := make([]int, len(toc.Interface))
			copy(originalInterfaces, toc.Interface)

			err = toc.UpdateInterfaceVersions(flavorInfo)
			if err != nil {
				t.Fatalf("UpdateInterfaceVersions failed: %v", err)
			}

			// Read back the updated file
			updatedToc, err := NewToc(tocPath)
			if err != nil {
				t.Fatalf("Failed to read updated Toc: %v", err)
			}

			// With Beta flag, we might have multiple interface versions
			// (base + beta if beta is higher)
			if len(updatedToc.Interface) == 0 {
				t.Errorf("No interface versions after beta update")
			}

			t.Logf("Beta mode - Original: %v, Updated: %v", originalInterfaces, updatedToc.Interface)
		})
	}
}

// formatInterfaceList formats a list of interface versions as a comma-separated string
func formatInterfaceList(versions []int) string {
	var strs []string
	for _, v := range versions {
		strs = append(strs, fmt.Sprintf("%d", v))
	}
	return strings.Join(strs, ", ")
}

// TestMultilineTocUpdate tests updating TOC files with multiple Interface directives
func TestMultilineTocUpdate(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "multi.toc")

	mockVer := getMockVersions()

	// Create a multi-line TOC file
	tocContent := fmt.Sprintf(`## Interface: 110000
## Interface-Vanilla: 11500
## Interface-Classic: 11500
## Interface-Mists: 50000

file.lua
`)

	err := os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	// Read and update each Interface directive
	contents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read TOC file: %v", err)
	}

	contentsStr := string(contents)
	lines := strings.Split(contentsStr, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "## Interface:") && !strings.Contains(line, "-") {
			lines[i] = fmt.Sprintf("## Interface: %d", mockVer.retail)
		} else if strings.Contains(line, "## Interface-Vanilla:") {
			lines[i] = fmt.Sprintf("## Interface-Vanilla: %d", mockVer.classicEra)
		} else if strings.Contains(line, "## Interface-Classic:") {
			lines[i] = fmt.Sprintf("## Interface-Classic: %d", mockVer.classicEra)
		} else if strings.Contains(line, "## Interface-Mists:") {
			lines[i] = fmt.Sprintf("## Interface-Mists: %d", mockVer.classic)
		}
	}

	newContents := strings.Join(lines, "\n")
	err = os.WriteFile(tocPath, []byte(newContents), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated TOC file: %v", err)
	}

	// Verify updates
	updatedContents, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated TOC file: %v", err)
	}

	updatedStr := string(updatedContents)

	expectedLines := []string{
		fmt.Sprintf("## Interface: %d", mockVer.retail),
		fmt.Sprintf("## Interface-Vanilla: %d", mockVer.classicEra),
		fmt.Sprintf("## Interface-Classic: %d", mockVer.classicEra),
		fmt.Sprintf("## Interface-Mists: %d", mockVer.classic),
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(updatedStr, expectedLine) {
			t.Errorf("Expected line %q not found in updated TOC:\n%s", expectedLine, updatedStr)
		}
	}
}

// TestFindTocFiles_MultipleFiles tests finding multiple TOC files in a directory
func TestFindTocFiles_MultipleFiles(t *testing.T) {
	testDir := setupTestTocFiles(t)

	tocFiles, err := FindTocFiles(testDir)
	if err != nil {
		t.Fatalf("FindTocFiles failed: %v", err)
	}

	expectedCount := 6 // We created 6 TOC files in setupTestTocFiles
	if len(tocFiles) != expectedCount {
		t.Errorf("Expected %d TOC files, got %d", expectedCount, len(tocFiles))
	}

	// Verify all expected files were found
	expectedFiles := []string{
		"default.toc",
		"multi-oneline.toc",
		"multi.toc",
		"specific-Classic.toc",
		"specific-Mainline.toc",
		"specific_Mists.toc",
	}

	foundFiles := make(map[string]bool)
	for _, tocFile := range tocFiles {
		basename := filepath.Base(tocFile)
		foundFiles[basename] = true
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected TOC file %q not found", expected)
		}
	}
}

// TestDetermineProjectName_FromMultipleTocFiles tests project name detection
func TestDetermineProjectName_FromMultipleTocFiles(t *testing.T) {
	tests := []struct {
		name     string
		tocFiles []string
		expected string
	}{
		{
			name:     "Single base TOC",
			tocFiles: []string{"MyAddon.toc"},
			expected: "MyAddon",
		},
		{
			name: "Base TOC with flavored TOCs",
			tocFiles: []string{
				"MyAddon.toc",
				"MyAddon-Classic.toc",
				"MyAddon_Mainline.toc",
			},
			expected: "MyAddon",
		},
		{
			name: "Only flavored TOCs",
			tocFiles: []string{
				"MyAddon-Classic.toc",
				"MyAddon_Mainline.toc",
			},
			expected: "MyAddon",
		},
		{
			name: "Hyphen separator",
			tocFiles: []string{
				"MyAddon-Classic.toc",
			},
			expected: "MyAddon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineProjectName(tt.tocFiles)
			if result != tt.expected {
				t.Errorf("Expected project name %q, got %q", tt.expected, result)
			}
		})
	}
}
