package toc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLatestBuildInfo(t *testing.T) {
	builds, err := GetLatestBuildInfo()
	assert.NoError(t, err, "GetLatestBuildInfo should not return an error")
	assert.NotNil(t, builds, "Expected builds to be non-nil")

	// Check for at least one product and one build info
	assert.Greater(t, len(*builds), 0, "Expected at least one product in builds")

	var products = []Product{}
	for product, buildInfos := range *builds {
		products = append(products, product)
		assert.NotEmpty(t, buildInfos, "Expected non-empty BuildInfo for product %s", product)
	}

	assert.Contains(t, products, ProductWow, "Expected ProductWow to be in the builds")
	assert.Contains(t, products, ProductWowBeta, "Expected ProductWowBeta to be in the builds")
	assert.Contains(t, products, ProductWowTest, "Expected ProductWowTest to be in the builds")
	assert.Contains(t, products, ProductWowXPtr, "Expected ProductWowXPtr to be in the builds")
	assert.Contains(t, products, ProductWowClassic, "Expected ProductWowClassic to be in the builds")
	assert.Contains(t, products, ProductWowClassicBeta, "Expected ProductWowClassicBeta to be in the builds")
	assert.Contains(t, products, ProductWowClassicPtr, "Expected ProductWowClassicPtr to be in the builds")
	assert.Contains(t, products, ProductWowClassicEra, "Expected ProductWowClassicEra to be in the builds")
	assert.Contains(t, products, ProductWowClassicEraPtr, "Expected ProductWowClassicEraPtr to be in the builds")
}
