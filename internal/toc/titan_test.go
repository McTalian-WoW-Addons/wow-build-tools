package toc

import (
	"slices"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/flavor"
)

// Titan (3.80.x) shares a major version with Wrath Classic (3.4.x), the same
// shape as Forever vs Classic Era on the 1.x line.
func TestGetFlavorFromVersion_TitanVsWrath(t *testing.T) {
	tests := []struct {
		major    int
		minor    int
		expected GameFlavor
	}{
		{3, 4, WotlkClassic},
		{3, 79, WotlkClassic},
		{3, 80, TitanClassic},
		{3, 81, TitanClassic},
	}

	for _, tt := range tests {
		if actual := getFlavorFromVersion(tt.major, tt.minor); actual != tt.expected {
			t.Errorf("getFlavorFromVersion(%d, %d) = %s, expected %s", tt.major, tt.minor, actual.ToString(), tt.expected.ToString())
		}
	}
}

func TestGetFlavorFromInterfaceVersion_Titan(t *testing.T) {
	if actual := getFlavorFromInterfaceVersion(30405); actual != WotlkClassic {
		t.Errorf("30405 = %s, expected wrath", actual.ToString())
	}
	if actual := getFlavorFromInterfaceVersion(38002); actual != TitanClassic {
		t.Errorf("38002 = %s, expected titan", actual.ToString())
	}
}

func TestTitanNaming(t *testing.T) {
	if TitanClassic.ToString() != "titan" {
		t.Errorf("Expected titan, got %s", TitanClassic.ToString())
	}
	if TitanClassic.DisplayName() != "Titan Classic" {
		t.Errorf("Expected Titan Classic, got %s", TitanClassic.DisplayName())
	}
	if TitanClassic.Label() != "Titan Classic (titan)" {
		t.Errorf("Expected Titan Classic (titan), got %s", TitanClassic.Label())
	}
}

func TestTocFileToGameFlavor_Titan(t *testing.T) {
	for _, noExt := range []string{"TestAddon-Titan", "TestAddon_Titan"} {
		if flavor, _ := TocFileToGameFlavor(noExt); flavor != TitanClassic {
			t.Errorf("TocFileToGameFlavor(%q) = %s, expected titan", noExt, flavor.ToString())
		}
	}
}

func TestCompatibleInstallFlavors_Titan(t *testing.T) {
	ids := []string{}
	for _, f := range CompatibleInstallFlavorsFromInterfaces([]int{38002}) {
		ids = append(ids, f.Id)
	}

	if !slices.Contains(ids, "classicTitan") {
		t.Errorf("Expected classicTitan, got %v", ids)
	}
	// 3.80.x must not be mistaken for the Wrath-era progression client.
	if slices.Contains(ids, "classic") {
		t.Errorf("Titan interface should not map to the classic install flavor, got %v", ids)
	}
}

func TestParseGameVersionSegment_Titan(t *testing.T) {
	origVersions, origInterfaces := gameVersions, gameInterfaces
	gameVersions, gameInterfaces = make(GameVersions), make(GameInterfaces)
	t.Cleanup(func() { gameVersions, gameInterfaces = origVersions, origInterfaces })

	if err := parseGameVersionSegment("3.80.2"); err != nil {
		t.Fatalf("parseGameVersionSegment failed: %v", err)
	}

	if !slices.Contains(gameVersions[TitanClassic], "3.80.2") {
		t.Errorf("Expected 3.80.2 under Titan, got %v", gameVersions)
	}
	if !slices.Contains(gameInterfaces[TitanClassic], 38002) {
		t.Errorf("Expected interface 38002 under Titan, got %v", gameInterfaces)
	}
	if len(gameVersions[WotlkClassic]) != 0 {
		t.Errorf("Expected nothing under Wrath, got %v", gameVersions[WotlkClassic])
	}

	if err := parseGameVersionSegment("titan"); err != nil {
		t.Errorf("Expected titan to be a valid game version argument: %v", err)
	}
}

func TestTitanProductMapping(t *testing.T) {
	if ProductToFlavorMap[ProductWowClassicTitan] != TitanClassic {
		t.Error("Expected wow_classic_titan to map to Titan")
	}
	if !slices.Contains(FlavorReleaseToProductMap[TitanFlavorRelease], ProductWowClassicTitan) {
		t.Error("Expected the Titan live release to check wow_classic_titan")
	}
}

// Install directories confirmed from each product's TACT product config
// (shared_container_default_subfolder).
func TestConfirmedInstallDirs(t *testing.T) {
	expected := map[string]string{
		"_retail_":          "retail",
		"_beta_":            "beta",
		"_ptr_":             "ptr",
		"_xptr_":            "xptr",
		"_classic_":         "classic",
		"_classic_ptr_":     "classicPtr",
		"_classic_beta_":    "classicBeta",
		"_classic_era_":     "classicEra",
		"_classic_era_ptr_": "classicEraPtr",
		"_anniversary_":     "anniversary",
		"_classic_titan_":   "classicTitan",
		"_dark_realm_":      "darkRealm",
		"_submission_":      "submission",
	}

	for dir, id := range expected {
		if actual := flavor.FromDir(dir); actual.Id != id {
			t.Errorf("FromDir(%q) = %q, expected %q", dir, actual.Id, id)
		}
	}
}
