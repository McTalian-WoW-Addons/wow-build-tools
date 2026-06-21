package toc

import (
	"slices"

	"github.com/McTalian/wow-build-tools/internal/flavor"
)

// CompatibleInstallFlavorsFromInterfaces maps interface versions to the rough
// set of installed WoW clients that can load an addon. Phase 1 intentionally
// collapses release channels into their broader client family.
func CompatibleInstallFlavorsFromInterfaces(interfaceVersions []int) []flavor.Flavor {
	compatibleFlavors := []flavor.Flavor{}
	seenFlavorIds := make(map[string]bool)

	for _, interfaceVersion := range interfaceVersions {
		gameFlavor := getFlavorFromMajorVersion(interfaceVersion / 10000)
		for _, installFlavor := range compatibleInstallFlavorsForGameFlavor(gameFlavor) {
			if installFlavor.IsUnknown() || seenFlavorIds[installFlavor.Id] {
				continue
			}

			compatibleFlavors = append(compatibleFlavors, installFlavor)
			seenFlavorIds[installFlavor.Id] = true
		}
	}

	return compatibleFlavors
}

func compatibleInstallFlavorsForGameFlavor(gameFlavor GameFlavor) []flavor.Flavor {
	switch gameFlavor {
	case Retail:
		return []flavor.Flavor{
			flavor.FromId("retail"),
			flavor.FromId("beta"),
			flavor.FromId("ptr"),
			flavor.FromId("xptr"),
		}
	case ClassicEra:
		return []flavor.Flavor{
			flavor.FromId("classicEra"),
			flavor.FromId("classicEraPtr"),
		}
	case CurrentAnniversary:
		return []flavor.Flavor{
			flavor.FromId("anniversary"),
		}
	case Unknown:
		return nil
	default:
		return []flavor.Flavor{
			flavor.FromId("classic"),
			flavor.FromId("classicBeta"),
			flavor.FromId("classicPtr"),
		}
	}
}

func (t *Toc) CompatibleInstallFlavors() []flavor.Flavor {
	compatibleFlavors := CompatibleInstallFlavorsFromInterfaces(t.Interface)
	slices.SortFunc(compatibleFlavors, func(a flavor.Flavor, b flavor.Flavor) int {
		if a.Id < b.Id {
			return -1
		}
		if a.Id > b.Id {
			return 1
		}
		return 0
	})

	return compatibleFlavors
}
