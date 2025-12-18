package toc_test

import (
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
