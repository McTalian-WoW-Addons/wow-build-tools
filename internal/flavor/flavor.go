package flavor

type Flavor string

const (
	Retail        Flavor = "retail"
	Classic       Flavor = "classic"
	ClassicEra    Flavor = "classicEra"
	Ptr           Flavor = "ptr"
	Xptr          Flavor = "xptr"
	ClassicPtr    Flavor = "classicPtr"
	ClassicEraPtr Flavor = "classicEraPtr"
	ClassicBeta   Flavor = "classicBeta"
)

type FlavorDir string

const (
	unknownDir       FlavorDir = ""
	retailDir        FlavorDir = "_retail_"
	classicDir       FlavorDir = "_classic_"
	classicEraDir    FlavorDir = "_classic_era_"
	ptrDir           FlavorDir = "_ptr_"
	xptrDir          FlavorDir = "_xptr_"
	classicPtrDir    FlavorDir = "_classic_ptr_"
	classicEraPtrDir FlavorDir = "_classic_era_ptr_"
	classicBetaDir   FlavorDir = "_classic_beta_"
)

var KnownFlavors = []Flavor{Retail, Classic, ClassicEra, Ptr, Xptr, ClassicPtr, ClassicEraPtr, ClassicBeta}

func (f Flavor) ToDir() string {
	switch f {
	case Retail:
		return string(retailDir)
	case Classic:
		return string(classicDir)
	case ClassicEra:
		return string(classicEraDir)
	case Ptr:
		return string(ptrDir)
	case Xptr:
		return string(xptrDir)
	case ClassicPtr:
		return string(classicPtrDir)
	case ClassicEraPtr:
		return string(classicEraPtrDir)
	case ClassicBeta:
		return string(classicBetaDir)
	default:
		return string(unknownDir)
	}
}

// StringToFlavor converts a string to a Flavor type
func StringToFlavor(s string) Flavor {
	switch s {
	case "retail":
		return Retail
	case "classic":
		return Classic
	case "classicera":
		return ClassicEra
	case "ptr":
		return Ptr
	case "xptr":
		return Xptr
	case "classicptr":
		return ClassicPtr
	case "classiceraptr":
		return ClassicEraPtr
	case "classicbeta":
		return ClassicBeta
	default:
		return Retail // Default to retail as a fallback
	}
}
