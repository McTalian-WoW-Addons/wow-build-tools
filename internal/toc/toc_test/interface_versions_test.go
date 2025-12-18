package toc_test

import (
	"slices"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
)

func TestGetLatestBuildInfo(t *testing.T) {
	builds, err := toc.GetLatestBuildInfo()
	assert.NoError(t, err, "GetLatestBuildInfo should not return an error")
	assert.NotNil(t, builds, "Expected builds to be non-nil")

	// Check for at least one product and one build info
	assert.Greater(t, len(*builds), 0, "Expected at least one product in builds")

	var products []toc.Product = []toc.Product{}
	for product, buildInfos := range *builds {
		products = append(products, product)
		assert.NotEmpty(t, buildInfos, "Expected non-empty BuildInfo for product %s", product)
	}

	assert.Contains(t, products, toc.ProductWow, "Expected ProductWow to be in the builds")
	assert.Contains(t, products, toc.ProductWowBeta, "Expected ProductWowBeta to be in the builds")
	assert.Contains(t, products, toc.ProductWowTest, "Expected ProductWowTest to be in the builds")
	assert.Contains(t, products, toc.ProductWowXPtr, "Expected ProductWowXPtr to be in the builds")
	assert.Contains(t, products, toc.ProductWowClassic, "Expected ProductWowClassic to be in the builds")
	assert.Contains(t, products, toc.ProductWowClassicBeta, "Expected ProductWowClassicBeta to be in the builds")
	assert.Contains(t, products, toc.ProductWowClassicPtr, "Expected ProductWowClassicPtr to be in the builds")
	assert.Contains(t, products, toc.ProductWowClassicEra, "Expected ProductWowClassicEra to be in the builds")
	assert.Contains(t, products, toc.ProductWowClassicEraPtr, "Expected ProductWowClassicEraPtr to be in the builds")
}

func TestCheckForInterfaceBumpsNormal(t *testing.T) {
	flavorReleaseInfo := toc.FlavorReleaseInfo{
		IsBeta: false,
		IsTest: false,
	}

	toc.AddGameInterface(toc.Retail, 110000)
	toc.AddGameInterface(toc.CurrentClassic, 50500)
	toc.AddGameInterface(toc.ClassicEra, 11000)

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	assert.NoError(t, err, "CheckForInterfaceBumps should not return an error")
	assert.Greater(t, len(availableInterfaces), 0, "Expected at least one available interface version")
	assert.Equal(t, 3, len(availableInterfaces), "Expected exactly three available interface versions")

	slices.Sort(availableInterfaces)
	assert.Greater(t, availableInterfaces[0], 11000, "Expected available interface version for classic era to be greater than 11000")
	assert.Greater(t, availableInterfaces[1], 50500, "Expected available interface version for classic to be greater than 50500")
	assert.Greater(t, availableInterfaces[2], 110000, "Expected available interface version for retail to be greater than 110000")
}

func TestCheckForInterfaceBumpsPtr(t *testing.T) {
	flavorReleaseInfo := toc.FlavorReleaseInfo{
		IsBeta: false,
		IsTest: true,
	}

	toc.AddGameInterface(toc.Retail, 110205)
	toc.AddGameInterface(toc.CurrentClassic, 50010)
	toc.AddGameInterface(toc.ClassicEra, 11010)

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	assert.NoError(t, err, "CheckForInterfaceBumps should not return an error")
	assert.Greater(t, len(availableInterfaces), 0, "Expected at least one available interface version")

	var iFaceMap = make(map[int]bool)
	for _, iface := range availableInterfaces {
		iFaceMap[iface] = true
	}

	assert.False(t, iFaceMap[110205], "Did not expect seeded retail interface version to be included")
	assert.False(t, iFaceMap[50010], "Did not expect seeded classic interface version to be included")
	assert.False(t, iFaceMap[11010], "Did not expect seeded classic era interface version to be included")

	var uniqIfaceCount int = 0
	for _, included := range iFaceMap {
		if included {
			uniqIfaceCount++
		}
	}

	assert.Greater(t, uniqIfaceCount, 3, "Expected at least one test release type")
}

func TestCheckForInterfaceBumpsBeta(t *testing.T) {
	flavorReleaseInfo := toc.FlavorReleaseInfo{
		IsBeta: true,
		IsTest: false,
	}

	toc.AddGameInterface(toc.Retail, 110205)
	toc.AddGameInterface(toc.CurrentClassic, 50010)
	toc.AddGameInterface(toc.ClassicEra, 11010)

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	assert.NoError(t, err, "CheckForInterfaceBumps should not return an error")
	assert.Greater(t, len(availableInterfaces), 0, "Expected at least one available interface version")

	var iFaceMap = make(map[int]bool)
	for _, iface := range availableInterfaces {
		iFaceMap[iface] = true
	}

	assert.False(t, iFaceMap[110205], "Did not expect seeded retail interface version to be included")
	assert.False(t, iFaceMap[50010], "Did not expect seeded classic interface version to be included")
	assert.False(t, iFaceMap[11010], "Did not expect seeded classic era interface version to be included")

	var uniqIfaceCount int = 0
	for _, included := range iFaceMap {
		if included {
			uniqIfaceCount++
		}
	}

	assert.Greater(t, uniqIfaceCount, 3, "Expected at least one beta release type")
}

func TestCheckForInterfaceBumpsBetaAndPtr(t *testing.T) {
	flavorReleaseInfo := toc.FlavorReleaseInfo{
		IsBeta: true,
		IsTest: true,
	}

	toc.AddGameInterface(toc.Retail, 110205)
	toc.AddGameInterface(toc.CurrentClassic, 50010)
	toc.AddGameInterface(toc.ClassicEra, 11010)

	availableInterfaces, err := toc.CheckForInterfaceBumps(flavorReleaseInfo)
	assert.NoError(t, err, "CheckForInterfaceBumps should not return an error")
	assert.Greater(t, len(availableInterfaces), 0, "Expected at least one available interface version")

	var iFaceMap = make(map[int]bool)
	for _, iface := range availableInterfaces {
		iFaceMap[iface] = true
	}

	assert.False(t, iFaceMap[110205], "Did not expect seeded retail interface version to be included")
	assert.False(t, iFaceMap[50010], "Did not expect seeded classic interface version to be included")
	assert.False(t, iFaceMap[11010], "Did not expect seeded classic era interface version to be included")

	var uniqIfaceCount int = 0
	for _, included := range iFaceMap {
		if included {
			uniqIfaceCount++
		}
	}

	assert.Greater(t, uniqIfaceCount, 5, "Expected at least two beta or PTR release types")
}
