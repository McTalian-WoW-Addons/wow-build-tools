package flavor

type Flavor struct {
	Id     string
	Name   string
	SubDir string
}

var KnownFlavors = []Flavor{
	{Id: "retail", Name: "Retail", SubDir: "_retail_"},
	{Id: "beta", Name: "Beta", SubDir: "_beta_"},
	{Id: "classic", Name: "Classic", SubDir: "_classic_"},
	{Id: "classicEra", Name: "Classic Era", SubDir: "_classic_era_"},
	{Id: "ptr", Name: "PTR", SubDir: "_ptr_"},
	{Id: "xptr", Name: "XPTR", SubDir: "_xptr_"},
	{Id: "classicPtr", Name: "Classic PTR", SubDir: "_classic_ptr_"},
	{Id: "classicEraPtr", Name: "Classic Era PTR", SubDir: "_classic_era_ptr_"},
	{Id: "classicBeta", Name: "Classic Beta", SubDir: "_classic_beta_"},
}
var UnknownFlavor = Flavor{Id: "unknown", Name: "Unknown", SubDir: ""}

func (f Flavor) IsUnknown() bool {
	return f.Id == UnknownFlavor.Id
}

var IdFlavorMap = map[string]Flavor{}
var DirFlavorMap = map[string]Flavor{}

func init() {
	for _, f := range KnownFlavors {
		IdFlavorMap[f.Id] = f
		DirFlavorMap[f.SubDir] = f
	}
}

func FromDir(dir string) Flavor {
	if f, ok := DirFlavorMap[dir]; ok {
		return f
	}
	return UnknownFlavor
}

func FromId(id string) Flavor {
	if f, ok := IdFlavorMap[id]; ok {
		return f
	}
	return UnknownFlavor
}
