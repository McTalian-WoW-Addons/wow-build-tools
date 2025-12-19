package github

import (
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

// Client defines the interface for GitHub API operations
type Client interface {
	// IsTokenSet returns true if GITHUB_OAUTH is set
	IsTokenSet() bool

	// GetRelease retrieves a release by slug and tag
	GetRelease(slug, tag string) (*GitHubRelease, error)

	// CreateRelease creates a new release
	CreateRelease(slug string, payload GitHubReleasePayload) (*GitHubRelease, error)

	// GetReleaseMetadataContents generates release.json content
	GetReleaseMetadataContents(name, version string, gameInterfaces toc.GameInterfaces, zipPaths ...string) (string, error)
}

// Release defines the interface for release operations
type Release interface {
	UpdateRelease(payload GitHubReleasePayload) error
	UploadAsset(fileName, filePath string, logGroup *logger.LogGroup) error
}

// DefaultClient implements Client using the actual GitHub API
type DefaultClient struct{}

func (c *DefaultClient) IsTokenSet() bool {
	return IsTokenSet()
}

func (c *DefaultClient) GetRelease(slug, tag string) (*GitHubRelease, error) {
	return GetRelease(slug, tag)
}

func (c *DefaultClient) CreateRelease(slug string, payload GitHubReleasePayload) (*GitHubRelease, error) {
	return CreateRelease(slug, payload)
}

func (c *DefaultClient) GetReleaseMetadataContents(name, version string, gameInterfaces toc.GameInterfaces, zipPaths ...string) (string, error) {
	return GetReleaseMetadataContents(name, version, gameInterfaces, zipPaths...)
}

// NewDefaultClient creates a new default GitHub client
func NewDefaultClient() Client {
	return &DefaultClient{}
}
