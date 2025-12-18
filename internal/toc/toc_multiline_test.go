package toc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiLineInterfaceUpdate(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "test.toc")

	// Create a TOC file with multiple Interface directives
	originalContent := `## Interface: 110007
## Interface-Vanilla: 11505
## Interface-Classic: 50000
## Interface-Mists: 50000

file.lua
`

	err := os.WriteFile(tocPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	// Parse the TOC
	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("Failed to parse TOC: %v", err)
	}

	t.Logf("Parsed interfaces: %v", toc.Interface)

	// Now update it
	flavorInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: false,
	}

	err = toc.UpdateInterfaceVersions(flavorInfo)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	// Read the file contents to see what was written
	updatedBytes, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedContent := string(updatedBytes)
	t.Logf("Updated file content:\n%s", updatedContent)

	// Parse it again to see what interfaces we get
	toc2, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("Failed to parse updated TOC: %v", err)
	}

	t.Logf("Re-parsed interfaces: %v", toc2.Interface)

	// Verify the file structure
	lines := strings.Split(updatedContent, "\n")
	interfaceLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "## Interface") {
			interfaceLines = append(interfaceLines, strings.TrimSpace(line))
		}
	}

	t.Logf("Interface lines in updated file: %v", interfaceLines)

	// Check that we have the expected number of Interface lines
	if len(interfaceLines) != 4 {
		t.Errorf("Expected 4 Interface lines, got %d: %v", len(interfaceLines), interfaceLines)
	}
}

func TestMultiLineInterfaceWithInactiveFlavor(t *testing.T) {
	tempDir := t.TempDir()
	tocPath := filepath.Join(tempDir, "test.toc")

	// Create a TOC file with an inactive flavor (Wrath)
	originalContent := `## Interface: 110007
## Interface-Vanilla: 11505
## Interface-Wrath: 30403
## Interface-Mists: 50000

file.lua
`

	err := os.WriteFile(tocPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test TOC file: %v", err)
	}

	// Parse the TOC
	toc, err := NewToc(tocPath)
	if err != nil {
		t.Fatalf("Failed to parse TOC: %v", err)
	}

	t.Logf("Parsed interfaces: %v", toc.Interface)

	// Update it
	flavorInfo := FlavorReleaseInfo{
		IsBeta: false,
		IsTest: false,
	}

	err = toc.UpdateInterfaceVersions(flavorInfo)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	// Read the file contents
	updatedBytes, err := os.ReadFile(tocPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedContent := string(updatedBytes)
	t.Logf("Updated file content:\n%s", updatedContent)

	// Verify that the Wrath line was NOT changed
	if !strings.Contains(updatedContent, "## Interface-Wrath: 30403") {
		t.Errorf("Expected Interface-Wrath line to remain unchanged at 30403, but it was modified:\n%s", updatedContent)
	}

	// Verify that other lines were updated
	if strings.Contains(updatedContent, "## Interface-Vanilla: 11505") {
		t.Errorf("Expected Interface-Vanilla to be updated from 11505")
	}
	if strings.Contains(updatedContent, "## Interface-Mists: 50000") {
		t.Errorf("Expected Interface-Mists to be updated from 50000")
	}
}
