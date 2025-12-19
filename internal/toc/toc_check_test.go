package toc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/cmdargs"
)

func TestDifference(t *testing.T) {
	tests := []struct {
		name      string
		a         []string
		b         []string
		wantMissA []string
		wantMissB []string
	}{
		{
			name:      "empty slices",
			a:         []string{},
			b:         []string{},
			wantMissA: []string{},
			wantMissB: []string{},
		},
		{
			name:      "identical slices",
			a:         []string{"file1.lua", "file2.lua"},
			b:         []string{"file1.lua", "file2.lua"},
			wantMissA: []string{},
			wantMissB: []string{},
		},
		{
			name:      "a has extra files",
			a:         []string{"file1.lua", "file2.lua", "file3.lua"},
			b:         []string{"file1.lua", "file2.lua"},
			wantMissA: []string{},
			wantMissB: []string{"file3.lua"},
		},
		{
			name:      "b has extra files",
			a:         []string{"file1.lua", "file2.lua"},
			b:         []string{"file1.lua", "file2.lua", "file3.lua"},
			wantMissA: []string{"file3.lua"},
			wantMissB: []string{},
		},
		{
			name:      "completely different",
			a:         []string{"file1.lua", "file2.lua"},
			b:         []string{"file3.lua", "file4.lua"},
			wantMissA: []string{"file3.lua", "file4.lua"},
			wantMissB: []string{"file1.lua", "file2.lua"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMissA, gotMissB := difference(tt.a, tt.b)

			if len(gotMissA) != len(tt.wantMissA) {
				t.Errorf("difference() missingA length = %v, want %v", len(gotMissA), len(tt.wantMissA))
			}
			if len(gotMissB) != len(tt.wantMissB) {
				t.Errorf("difference() missingB length = %v, want %v", len(gotMissB), len(tt.wantMissB))
			}

			missAMap := make(map[string]bool)
			for _, v := range gotMissA {
				missAMap[v] = true
			}
			for _, v := range tt.wantMissA {
				if !missAMap[v] {
					t.Errorf("difference() missingA missing expected value %v", v)
				}
			}

			missBMap := make(map[string]bool)
			for _, v := range gotMissB {
				missBMap[v] = true
			}
			for _, v := range tt.wantMissB {
				if !missBMap[v] {
					t.Errorf("difference() missingB missing expected value %v", v)
				}
			}
		})
	}
}

func TestTocCheckParams(t *testing.T) {
	if TocCheckParams == nil {
		t.Fatal("TocCheckParams should not be nil")
	}

	if TocCheckParams.IgnoreFiles == nil {
		t.Error("IgnoreFiles should be initialized")
	}

	if TocCheckParams.SkipInterfaceCheck != false {
		t.Error("SkipInterfaceCheck should default to false")
	}

	if TocCheckParams.SkipMissingFilesCheck != false {
		t.Error("SkipMissingFilesCheck should default to false")
	}

	if TocCheckParams.SkipNameCheck != false {
		t.Error("SkipNameCheck should default to false")
	}
}

func TestRunTocCheck_NoTocFiles(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Save original params and restore after test
	origParams := TocParams
	defer func() { TocParams = origParams }()

	TocParams = &TocArgs{
		AddonDir: tempDir,
	}

	err := RunTocCheck()
	if err == nil {
		t.Error("Expected error when no TOC files found, got nil")
	}
}

func TestRunTocCheck_WithValidToc(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	addonName := "TestAddon"
	addonDir := filepath.Join(tempDir, addonName)
	err := os.Mkdir(addonDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create addon directory: %v", err)
	}

	// Create a simple TOC file
	tocPath := filepath.Join(addonDir, addonName+".toc")
	tocContent := `## Interface: 110002
## Title: Test Addon
## Version: 1.0.0

TestAddon.lua
`
	err = os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create TOC file: %v", err)
	}

	// Create corresponding lua file
	luaPath := filepath.Join(addonDir, "TestAddon.lua")
	err = os.WriteFile(luaPath, []byte("-- Test lua file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create lua file: %v", err)
	}

	// Save original params and restore after test
	origTocParams := TocParams
	origTocCheckParams := TocCheckParams
	origRootParams := cmdargs.RootParams
	defer func() {
		TocParams = origTocParams
		TocCheckParams = origTocCheckParams
		cmdargs.RootParams = origRootParams
	}()

	TocParams = &TocArgs{
		AddonDir: addonDir,
	}
	TocCheckParams = &TocCheckArgs{
		IgnoreFiles:           []string{},
		SkipInterfaceCheck:    true, // Skip to avoid external API calls
		SkipMissingFilesCheck: false,
		SkipNameCheck:         false,
	}
	cmdargs.RootParams = &cmdargs.RootArgs{
		LevelVerbose: false,
		LevelDebug:   false,
	}

	err = RunTocCheck()
	if err != nil {
		t.Errorf("Expected no error with valid TOC, got: %v", err)
	}
}

func TestRunTocCheck_WithRelativePath(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	addonName := "TestAddon"
	addonDir := filepath.Join(tempDir, addonName)
	err := os.Mkdir(addonDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create addon directory: %v", err)
	}

	// Create a simple TOC file
	tocPath := filepath.Join(addonDir, addonName+".toc")
	tocContent := `## Interface: 110002
## Title: Test Addon
## Version: 1.0.0

TestAddon.lua
`
	err = os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create TOC file: %v", err)
	}

	// Create corresponding lua file
	luaPath := filepath.Join(addonDir, "TestAddon.lua")
	err = os.WriteFile(luaPath, []byte("-- Test lua file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create lua file: %v", err)
	}

	// Change to temp directory to test relative paths
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		err := os.Chdir(origDir)
		if err != nil {
			t.Fatalf("Failed to change directory back to original: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Save original params and restore after test
	origTocParams := TocParams
	origTocCheckParams := TocCheckParams
	origRootParams := cmdargs.RootParams
	defer func() {
		TocParams = origTocParams
		TocCheckParams = origTocCheckParams
		cmdargs.RootParams = origRootParams
	}()

	// Use relative path
	TocParams = &TocArgs{
		AddonDir: addonName,
	}
	TocCheckParams = &TocCheckArgs{
		IgnoreFiles:           []string{},
		SkipInterfaceCheck:    true, // Skip to avoid external API calls
		SkipMissingFilesCheck: false,
		SkipNameCheck:         false,
	}
	cmdargs.RootParams = &cmdargs.RootArgs{
		LevelVerbose: false,
		LevelDebug:   false,
	}

	err = RunTocCheck()
	if err != nil {
		t.Errorf("Expected no error with valid TOC using relative path, got: %v", err)
	}
}

func TestRunTocCheck_WithSubdirectoryAndXml(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	addonName := "TestAddon"
	addonDir := filepath.Join(tempDir, addonName)
	err := os.Mkdir(addonDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create addon directory: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(addonDir, "Modules")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create TOC file that includes an XML file
	tocPath := filepath.Join(addonDir, addonName+".toc")
	tocContent := `## Interface: 110002
## Title: Test Addon
## Version: 1.0.0

TestAddon.lua
embed.xml
`
	err = os.WriteFile(tocPath, []byte(tocContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create TOC file: %v", err)
	}

	// Create main lua file
	luaPath := filepath.Join(addonDir, "TestAddon.lua")
	err = os.WriteFile(luaPath, []byte("-- Test lua file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main lua file: %v", err)
	}

	// Create XML file that includes a lua file in subdirectory
	xmlPath := filepath.Join(addonDir, "embed.xml")
	xmlContent := `<Ui xmlns="http://www.blizzard.com/wow/ui/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.blizzard.com/wow/ui/">
	<Script file="Modules\ModuleFile.lua"/>
</Ui>`
	err = os.WriteFile(xmlPath, []byte(xmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create XML file: %v", err)
	}

	// Create lua file in subdirectory
	moduleLuaPath := filepath.Join(subDir, "ModuleFile.lua")
	err = os.WriteFile(moduleLuaPath, []byte("-- Module lua file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create module lua file: %v", err)
	}

	// Save original params and restore after test
	origTocParams := TocParams
	origTocCheckParams := TocCheckParams
	origRootParams := cmdargs.RootParams
	defer func() {
		TocParams = origTocParams
		TocCheckParams = origTocCheckParams
		cmdargs.RootParams = origRootParams
	}()

	TocParams = &TocArgs{
		AddonDir: addonDir,
	}
	TocCheckParams = &TocCheckArgs{
		IgnoreFiles:           []string{},
		SkipInterfaceCheck:    true, // Skip to avoid external API calls
		SkipMissingFilesCheck: false,
		SkipNameCheck:         false,
	}
	cmdargs.RootParams = &cmdargs.RootArgs{
		LevelVerbose: false,
		LevelDebug:   false,
	}

	err = RunTocCheck()
	if err != nil {
		t.Errorf("Expected no error with valid TOC including XML with subdirectory lua file, got: %v", err)
	}
}
