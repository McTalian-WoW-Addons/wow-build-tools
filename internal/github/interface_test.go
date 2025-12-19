package github

import (
	"os"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
)

func TestDefaultClient_IsTokenSet(t *testing.T) {
	client := NewDefaultClient()

	// Save original
	origToken := os.Getenv("GITHUB_OAUTH")

	// Test with token
	_ = os.Setenv("GITHUB_OAUTH", "test-token")
	assert.True(t, client.IsTokenSet())

	// Test without token
	_ = os.Unsetenv("GITHUB_OAUTH")
	assert.False(t, client.IsTokenSet())

	// Restore
	if origToken != "" {
		_ = os.Setenv("GITHUB_OAUTH", origToken)
	} else {
		_ = os.Unsetenv("GITHUB_OAUTH")
	}
}

func TestDefaultClient_GetReleaseMetadataContents(t *testing.T) {
	client := NewDefaultClient()

	packageName := "TestAddon"
	version := "1.0.0"
	gameInterfaces := toc.GameInterfaces{
		toc.Retail: []int{110000},
	}
	zipPaths := []string{"TestAddon-1.0.0.zip"}

	result, err := client.GetReleaseMetadataContents(packageName, version, gameInterfaces, zipPaths...)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, packageName)
	assert.Contains(t, result, version)
}

func TestNewDefaultClient(t *testing.T) {
	client := NewDefaultClient()
	assert.NotNil(t, client)

	// Verify it implements the Client interface
	var _ = client
}
