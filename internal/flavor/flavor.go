package flavor

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
	{Id: "ptr", Name: "PTR", Dir: "_ptr_"},
	{Id: "xptr", Name: "XPTR", Dir: "_xptr_"},
	{Id: "classicPtr", Name: "Classic PTR", Dir: "_classic_ptr_"},
	{Id: "classicEraPtr", Name: "Classic Era PTR", Dir: "_classic_era_ptr_"},
	{Id: "classicBeta", Name: "Classic Beta", Dir: "_classic_beta_"},
}
var UnknownFlavor = Flavor{Id: "unknown", Name: "Unknown", Dir: ""}

func (f Flavor) IsUnknown() bool {
	return f.Id == UnknownFlavor.Id
}

var IdFlavorMap = map[string]Flavor{}
var DirFlavorMap = map[string]Flavor{}

func init() {
	for _, f := range KnownFlavors {
		IdFlavorMap[f.Id] = f
		DirFlavorMap[f.Dir] = f
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
