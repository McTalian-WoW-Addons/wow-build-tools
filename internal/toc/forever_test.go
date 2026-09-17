package toc

import (
	"slices"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/flavor"
)

// Forever (1.60.x) shares a major version with Classic Era (1.13-1.15), so the
// minor version is the only thing separating them.
func TestGetFlavorFromVersion_ForeverVsClassicEra(t *testing.T) {
	tests := []struct {
		major    int
		minor    int
		expected GameFlavor
	}{
		{1, 13, ClassicEra},
		{1, 15, ClassicEra},
		{1, 59, ClassicEra},
		{1, 60, Forever},
		{1, 61, Forever},
		{2, 5, TbcClassic},
		{5, 5, MistsClassic},
		{12, 1, Retail},
	}

	for _, tt := range tests {
		if actual := getFlavorFromVersion(tt.major, tt.minor); actual != tt.expected {
			t.Errorf("getFlavorFromVersion(%d, %d) = %s, expected %s", tt.major, tt.minor, actual.ToString(), tt.expected.ToString())
		}
	}
}

func TestGetFlavorFromInterfaceVersion(t *testing.T) {
	tests := []struct {
		interfaceVersion int
		expected         GameFlavor
	}{
		{11509, ClassicEra},
		{16001, Forever},
		{20506, TbcClassic},
		{50504, MistsClassic},
		{120100, Retail},
	}

	for _, tt := range tests {
		if actual := getFlavorFromInterfaceVersion(tt.interfaceVersion); actual != tt.expected {
			t.Errorf("getFlavorFromInterfaceVersion(%d) = %s, expected %s", tt.interfaceVersion, actual.ToString(), tt.expected.ToString())
		}
	}
}

func TestForeverToString(t *testing.T) {
	if Forever.ToString() != "forever" {
		t.Errorf("Expected forever, got %s", Forever.ToString())
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		flavor   GameFlavor
		expected string
	}{
		{ClassicEra, "Classic Era"},
		{TbcClassic, "Burning Crusade Classic"},
		{MistsClassic, "Mists Classic"},
		{Forever, "Forever"},
		{Retail, "Retail"},
	}

	for _, tt := range tests {
		if actual := tt.flavor.DisplayName(); actual != tt.expected {
			t.Errorf("DisplayName() = %q, expected %q", actual, tt.expected)
		}
	}

	// The slug and the display name are deliberately different things.
	if ClassicEra.ToString() == ClassicEra.DisplayName() {
		t.Error("Expected the Classic Era slug and display name to differ")
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		flavor   GameFlavor
		expected string
	}{
		{ClassicEra, "Classic Era (classic)"},
		{TbcClassic, "Burning Crusade Classic (bcc)"},
		{Forever, "Forever (forever)"},
		{Retail, "Retail (retail)"},
	}

	for _, tt := range tests {
		if actual := tt.flavor.Label(); actual != tt.expected {
			t.Errorf("Label() = %q, expected %q", actual, tt.expected)
		}
	}
}

func TestTocFileToGameFlavor_Forever(t *testing.T) {
	for _, noExt := range []string{"TestAddon-Forever", "TestAddon_Forever", "TestAddon-forever"} {
		if flavor, _ := TocFileToGameFlavor(noExt); flavor != Forever {
			t.Errorf("TocFileToGameFlavor(%q) = %s, expected forever", noExt, flavor.ToString())
		}
	}
}

// A Forever interface must not drag in the Classic Era install directories.
func TestCompatibleInstallFlavors_Forever(t *testing.T) {
	compatibleFlavors := CompatibleInstallFlavorsFromInterfaces([]int{16001})

	ids := []string{}
	for _, f := range compatibleFlavors {
		ids = append(ids, f.Id)
	}

	if !slices.Contains(ids, "forever") {
		t.Errorf("Expected forever install flavor, got %v", ids)
	}
	if slices.Contains(ids, "classicEra") {
		t.Errorf("Forever interface should not map to Classic Era, got %v", ids)
	}
	for _, f := range compatibleFlavors {
		if f.IsUnknown() {
			t.Error("Expected no unknown install flavors")
		}
	}
}

func TestForeverInstallFlavorsAreRegistered(t *testing.T) {
	for _, id := range []string{"forever", "foreverBeta"} {
		if flavor.FromId(id).IsUnknown() {
			t.Errorf("Install flavor %q is not registered", id)
		}
	}
	if flavor.FromDir("_forever_").Id != "forever" {
		t.Errorf("Expected _forever_ to resolve to forever, got %s", flavor.FromDir("_forever_").Id)
	}
}

func TestParseGameVersionSegment_Forever(t *testing.T) {
	origVersions, origInterfaces := gameVersions, gameInterfaces
	gameVersions, gameInterfaces = make(GameVersions), make(GameInterfaces)
	t.Cleanup(func() { gameVersions, gameInterfaces = origVersions, origInterfaces })

	if err := parseGameVersionSegment("1.60.1"); err != nil {
		t.Fatalf("parseGameVersionSegment failed: %v", err)
	}

	if !slices.Contains(gameVersions[Forever], "1.60.1") {
		t.Errorf("Expected 1.60.1 under Forever, got %v", gameVersions)
	}
	if !slices.Contains(gameInterfaces[Forever], 16001) {
		t.Errorf("Expected interface 16001 under Forever, got %v", gameInterfaces)
	}
	if len(gameVersions[ClassicEra]) != 0 {
		t.Errorf("Expected nothing under Classic Era, got %v", gameVersions[ClassicEra])
	}

	// The bare flavor name is accepted as a --game-version argument.
	if err := parseGameVersionSegment("forever"); err != nil {
		t.Errorf("Expected forever to be a valid game version argument: %v", err)
	}
}

func TestForeverFlavorReleasesHaveProducts(t *testing.T) {
	for _, release := range []GameFlavorRelease{ForeverFlavorRelease, ForeverBetaFlavorRelease, ForeverTestFlavorRelease} {
		if len(FlavorReleaseToProductMap[release]) == 0 {
			t.Errorf("No products mapped for %s", release.ToString())
		}
	}

	// The 1.60.x beta is currently served through wow_classic_beta.
	if !slices.Contains(FlavorReleaseToProductMap[ForeverBetaFlavorRelease], ProductWowClassicBeta) {
		t.Error("Expected wow_classic_beta to be checked for the Forever beta")
	}
}
