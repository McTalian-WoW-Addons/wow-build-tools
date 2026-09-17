package flavor

import "strings"

type Flavor struct {
	Id   string
	Name string
	Dir  string
}

var KnownFlavors = []Flavor{
	{Id: "retail", Name: "Retail", Dir: "_retail_"},
	{Id: "beta", Name: "Beta", Dir: "_beta_"},
	{Id: "classic", Name: "Classic", Dir: "_classic_"},
	{Id: "classicEra", Name: "Classic Era", Dir: "_classic_era_"},
	{Id: "anniversary", Name: "Classic Anniversary", Dir: "_anniversary_"},
	{Id: "ptr", Name: "PTR", Dir: "_ptr_"},
	{Id: "xptr", Name: "XPTR", Dir: "_xptr_"},
	{Id: "classicPtr", Name: "Classic PTR", Dir: "_classic_ptr_"},
	{Id: "classicEraPtr", Name: "Classic Era PTR", Dir: "_classic_era_ptr_"},
	{Id: "classicBeta", Name: "Classic Beta", Dir: "_classic_beta_"},
	{Id: "classicTitan", Name: "Classic Titan", Dir: "_classic_titan_"},
	// Internal Blizzard test clients. Listed so FromDir resolves them instead
	// of returning Unknown when config scans a WoW install.
	{Id: "darkRealm", Name: "Dark Realm", Dir: "_dark_realm_"},
	{Id: "submission", Name: "Submission", Dir: "_submission_"},
	// TODO(forever): Blizzard has not published a Forever product yet, so these
	// directories do not exist. The 1.60.x beta installs into _classic_beta_
	// (per the wow_classic_beta TACT product config), which is why Forever also
	// maps to the classicBeta install flavor. These follow the existing naming
	// convention and are inert until a matching directory exists on disk.
	{Id: "forever", Name: "Forever", Dir: "_forever_"},
	{Id: "foreverBeta", Name: "Forever Beta", Dir: "_forever_beta_"},
}
var UnknownFlavor = Flavor{Id: "unknown", Name: "Unknown", Dir: ""}

func (f Flavor) IsUnknown() bool {
	return f.Id == UnknownFlavor.Id
}

var IdFlavorMap = map[string]Flavor{}
var DirFlavorMap = map[string]Flavor{}

// Lookups are case-insensitive. Viper lowercases every configuration key, so a
// camelCase id such as "classicBeta" comes back out of the config as
// "classicbeta" and would otherwise never resolve.
var lowerIdFlavorMap = map[string]Flavor{}
var lowerDirFlavorMap = map[string]Flavor{}

func init() {
	for _, f := range KnownFlavors {
		IdFlavorMap[f.Id] = f
		DirFlavorMap[f.Dir] = f
		lowerIdFlavorMap[strings.ToLower(f.Id)] = f
		lowerDirFlavorMap[strings.ToLower(f.Dir)] = f
	}
}

func FromDir(dir string) Flavor {
	if f, ok := lowerDirFlavorMap[strings.ToLower(dir)]; ok {
		return f
	}
	return UnknownFlavor
}

func FromId(id string) Flavor {
	if f, ok := lowerIdFlavorMap[strings.ToLower(id)]; ok {
		return f
	}
	return UnknownFlavor
}
