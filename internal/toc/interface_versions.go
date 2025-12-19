package toc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type BuildInfo struct {
	Product       string `json:"product"`
	Version       string `json:"version"`
	CreatedAt     string `json:"created_at"`
	BuildConfig   string `json:"build_config"`
	ProductConfig string `json:"product_config"`
	CdnConfig     string `json:"cdn_config"`
}

type Product string

const (
	ProductWow              Product = "wow"
	ProductWowBeta          Product = "wow_beta"
	ProductWowTest          Product = "wowt"
	ProductWowXPtr          Product = "wowxptr"
	ProductWowLiveTest      Product = "wowlivetest"
	ProductWowZ             Product = "wowz"
	ProductWowClassic       Product = "wow_classic"
	ProductWowClassicBeta   Product = "wow_classic_beta"
	ProductWowClassicPtr    Product = "wow_classic_ptr"
	ProductWoWClassicTitan  Product = "wow_classic_titan"
	ProductWowClassicEra    Product = "wow_classic_era"
	ProductWowClassicEraPtr Product = "wow_classic_era_ptr"
)

var ProductToFlavorMap map[Product]GameFlavor = map[Product]GameFlavor{
	ProductWow:              Retail,
	ProductWowBeta:          Retail,
	ProductWowTest:          Retail,
	ProductWowXPtr:          Retail,
	ProductWowLiveTest:      Retail, // Not sure about this one
	ProductWowZ:             Retail, // Not sure about this one
	ProductWowClassic:       CurrentClassic,
	ProductWowClassicBeta:   CurrentClassic,
	ProductWowClassicPtr:    CurrentClassic,
	ProductWoWClassicTitan:  WotlkClassic, // This one's a bit more nuanced than that
	ProductWowClassicEra:    ClassicEra,
	ProductWowClassicEraPtr: ClassicEra,
}

type FlavorReleaseInfo struct {
	IsBeta bool
	IsTest bool
}

type GameReleaseType int

const (
	LiveRelease GameReleaseType = iota
	BetaRelease
	TestRelease
)

func (gr GameReleaseType) ToString() string {
	switch gr {
	case LiveRelease:
		return "live"
	case BetaRelease:
		return "beta"
	case TestRelease:
		return "test"
	default:
		return "unknown"
	}
}

type GameFlavorRelease struct {
	Flavor      GameFlavor
	ReleaseType GameReleaseType
}

func (gr GameFlavorRelease) ToString() string {
	return fmt.Sprintf("%s-%s", gr.Flavor.ToString(), gr.ReleaseType.ToString())
}

var (
	RetailFlavorRelease     = GameFlavorRelease{Flavor: Retail, ReleaseType: LiveRelease}
	RetailBetaFlavorRelease = GameFlavorRelease{Flavor: Retail, ReleaseType: BetaRelease}
	RetailTestFlavorRelease = GameFlavorRelease{Flavor: Retail, ReleaseType: TestRelease}

	ClassicEraFlavorRelease     = GameFlavorRelease{Flavor: ClassicEra, ReleaseType: LiveRelease}
	ClassicEraBetaFlavorRelease = GameFlavorRelease{Flavor: ClassicEra, ReleaseType: BetaRelease}
	ClassicEraTestFlavorRelease = GameFlavorRelease{Flavor: ClassicEra, ReleaseType: TestRelease}

	ClassicFlavorRelease     = GameFlavorRelease{Flavor: CurrentClassic, ReleaseType: LiveRelease}
	ClassicBetaFlavorRelease = GameFlavorRelease{Flavor: CurrentClassic, ReleaseType: BetaRelease}
	ClassicTestFlavorRelease = GameFlavorRelease{Flavor: CurrentClassic, ReleaseType: TestRelease}
)

var FlavorReleaseToProductMap map[GameFlavorRelease][]Product = map[GameFlavorRelease][]Product{
	RetailFlavorRelease:     {ProductWow},
	RetailBetaFlavorRelease: {ProductWowBeta},
	RetailTestFlavorRelease: {ProductWowTest, ProductWowXPtr},

	ClassicEraFlavorRelease:     {ProductWowClassicEra},
	ClassicEraBetaFlavorRelease: {ProductWowClassicEraPtr}, // Classic Era Beta doesn't really show up in build info right now
	ClassicEraTestFlavorRelease: {ProductWowClassicEraPtr},

	ClassicFlavorRelease:     {ProductWowClassic},
	ClassicBetaFlavorRelease: {ProductWowClassicBeta},
	ClassicTestFlavorRelease: {ProductWowClassicPtr},
}

type ProductBuilds = map[Product]BuildInfo

var wagoApiUrl = "https://wago.tools/api"
var latestBuilds = fmt.Sprintf("%s/builds/latest", wagoApiUrl)

var cacheLatestBuilds *ProductBuilds = nil

func GetLatestBuildInfo() (*ProductBuilds, error) {
	// Return cached builds if available
	if cacheLatestBuilds != nil {
		return cacheLatestBuilds, nil
	}

	req, err := http.NewRequest("GET", latestBuilds, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var builds ProductBuilds
	err = json.NewDecoder(resp.Body).Decode(&builds)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	cacheLatestBuilds = &builds

	// Implementation to fetch and parse latest build info from Wago API
	return &builds, nil
}

func (b *BuildInfo) GetInterfaceVersion() (int, error) {
	segments := strings.Split(b.Version, ".")
	if len(segments) < 3 {
		return 0, fmt.Errorf("invalid build version format: %s", b.Version)
	}

	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return 0, fmt.Errorf("invalid major version in build version: %s", b.Version)
	}

	minor, err := strconv.Atoi(segments[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minor version in build version: %s", b.Version)
	}

	patch, err := strconv.Atoi(segments[2])
	if err != nil {
		return 0, fmt.Errorf("invalid patch version in build version: %s", b.Version)
	}

	interfaceVersion, err := strconv.Atoi(fmt.Sprintf("%d%02d%02d", major, minor, patch))
	if err != nil {
		return 0, fmt.Errorf("failed to construct interface version: %w", err)
	}

	return interfaceVersion, nil
}
