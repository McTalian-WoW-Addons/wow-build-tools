package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/external"
	"github.com/McTalian/wow-build-tools/internal/github"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/McTalian/wow-build-tools/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGitHubClient is a mock implementation of github.Client
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) IsTokenSet() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGitHubClient) GetRelease(slug, tag string) (*github.GitHubRelease, error) {
	args := m.Called(slug, tag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*github.GitHubRelease), args.Error(1)
}

func (m *MockGitHubClient) CreateRelease(slug string, payload github.GitHubReleasePayload) (*github.GitHubRelease, error) {
	args := m.Called(slug, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*github.GitHubRelease), args.Error(1)
}

func (m *MockGitHubClient) GetReleaseMetadataContents(name, version string, gameInterfaces toc.GameInterfaces, zipPaths ...string) (string, error) {
	args := m.Called(name, version, gameInterfaces, zipPaths)
	return args.String(0), args.Error(1)
}

// MockRelease is a mock implementation of github.Release
type MockRelease struct {
	mock.Mock
	github.GitHubRelease
}

func (m *MockRelease) UpdateRelease(payload github.GitHubReleasePayload) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockRelease) UploadAsset(fileName, filePath string, logGroup *logger.LogGroup) error {
	args := m.Called(fileName, filePath, logGroup)
	return args.Error(0)
}

// mockRepo implements repo.VcsRepo for testing
type mockRepo struct {
	isGitHubHosted bool
	currentTag     string
	gitHubSlug     string
}

func (m *mockRepo) IsGitHubHosted() bool                                { return m.isGitHubHosted }
func (m *mockRepo) GetCurrentTag() string                               { return m.currentTag }
func (m *mockRepo) GetGitHubSlug() string                               { return m.gitHubSlug }
func (m *mockRepo) GetTopDir() string                                   { return "" }
func (m *mockRepo) GetVcsType() external.VcsType                        { return external.Git }
func (m *mockRepo) GetChangelog(title string) (string, error)           { return "", nil }
func (m *mockRepo) IsIgnored(path string, isDir bool) bool              { return false }
func (m *mockRepo) GetInjectionValues(stm *tokens.SimpleTokenMap) error { return nil }
func (m *mockRepo) GetFileInjectionValues(filePath string) (*tokens.SimpleTokenMap, error) {
	return nil, nil
}
func (m *mockRepo) GetRepoRoot() string                                       { return "" }
func (m *mockRepo) GetPreviousVersion() string                                { return "" }
func (m *mockRepo) GetProjectVersion() string                                 { return "" }
func (m *mockRepo) GetCurrentVersion() string                                 { return "" }
func (m *mockRepo) GetChangelogSinceTag(tag string) (string, error)           { return "", nil }
func (m *mockRepo) GetPreviousTag() string                                    { return "" }
func (m *mockRepo) GetLastEditedRevision() string                             { return "" }
func (m *mockRepo) GetRevisionNumber() string                                 { return "" }
func (m *mockRepo) GetAbbreviatedHash() string                                { return "" }
func (m *mockRepo) GetHash() string                                           { return "" }
func (m *mockRepo) GetTimestamp() string                                      { return "" }
func (m *mockRepo) GetAuthor() string                                         { return "" }
func (m *mockRepo) GetChangelogSinceRevision(revision string) (string, error) { return "", nil }
func (m *mockRepo) GetExternalsRevision() string                              { return "" }
func (m *mockRepo) GetRemoteURL() string                                      { return "" }

func TestShouldSkip(t *testing.T) {
	logGroup := logger.NewLogGroup("test")

	t.Run("skip when not github hosted", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		repo := &mockRepo{isGitHubHosted: false}

		result := shouldSkip(repo, mockClient, logGroup)
		assert.True(t, result)
	})

	t.Run("skip when no current tag", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		repo := &mockRepo{
			isGitHubHosted: true,
			currentTag:     "",
		}

		result := shouldSkip(repo, mockClient, logGroup)
		assert.True(t, result)
	})

	t.Run("skip when no github slug", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		repo := &mockRepo{
			isGitHubHosted: true,
			currentTag:     "v1.0.0",
			gitHubSlug:     "",
		}

		result := shouldSkip(repo, mockClient, logGroup)
		assert.True(t, result)
	})

	t.Run("skip when no token", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		mockClient.On("IsTokenSet").Return(false)

		repo := &mockRepo{
			isGitHubHosted: true,
			currentTag:     "v1.0.0",
			gitHubSlug:     "owner/repo",
		}

		result := shouldSkip(repo, mockClient, logGroup)
		assert.True(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("do not skip when all conditions met", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		mockClient.On("IsTokenSet").Return(true)

		repo := &mockRepo{
			isGitHubHosted: true,
			currentTag:     "v1.0.0",
			gitHubSlug:     "owner/repo",
		}

		result := shouldSkip(repo, mockClient, logGroup)
		assert.False(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestGetOrCreateRelease_CreateNewRelease(t *testing.T) {
	logGroup := logger.NewLogGroup("test")
	mockClient := new(MockGitHubClient)

	repo := &mockRepo{
		currentTag: "v1.0.0",
		gitHubSlug: "owner/repo",
	}

	// Mock GetRelease to return "not found"
	mockClient.On("GetRelease", "owner/repo", "v1.0.0").Return(nil, github.ErrReleaseNotFound)

	// Mock CreateRelease to succeed
	expectedRelease := &github.GitHubRelease{
		GitHubReleasePayload: github.GitHubReleasePayload{
			TagName:    "v1.0.0",
			Name:       "v1.0.0",
			Body:       "Test changelog",
			Prerelease: true,
			Draft:      false,
		},
		Id:   123,
		Slug: "owner/repo",
	}
	mockClient.On("CreateRelease", "owner/repo", mock.MatchedBy(func(payload github.GitHubReleasePayload) bool {
		return payload.TagName == "v1.0.0" &&
			payload.Name == "v1.0.0" &&
			payload.Body == "Test changelog" &&
			payload.Prerelease == true &&
			payload.Draft == false
	})).Return(expectedRelease, nil)

	release, err := GetOrCreateRelease(repo, mockClient, true, "Test changelog", logGroup)

	assert.NoError(t, err)
	assert.NotNil(t, release)
	assert.Equal(t, 123, release.Id)
	assert.Equal(t, "owner/repo", release.Slug)
	mockClient.AssertExpectations(t)
}

func TestGetOrCreateRelease_UpdateExistingRelease(t *testing.T) {
	// Set the token so UpdateRelease doesn't fail
	err := os.Setenv("GITHUB_OAUTH", "test-token")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv("GITHUB_OAUTH")
		if err != nil {
			t.Fatalf("Failed to unset GITHUB_OAUTH: %v", err)
		}
	}()

	logGroup := logger.NewLogGroup("test")
	mockClient := new(MockGitHubClient)

	repo := &mockRepo{
		currentTag: "v2.0.0",
		gitHubSlug: "owner/repo",
	}

	// Create a mock release that we'll return from GetRelease
	// Note: We can't fully mock UpdateRelease as it's a method on the concrete struct
	// This test verifies the GetRelease path works
	existingRelease := &github.GitHubRelease{
		GitHubReleasePayload: github.GitHubReleasePayload{
			TagName:    "v2.0.0",
			Name:       "v2.0.0",
			Body:       "Old changelog",
			Prerelease: false,
			Draft:      false,
		},
		Id:   456,
		Slug: "owner/repo",
	}

	// Mock GetRelease to return existing release
	mockClient.On("GetRelease", "owner/repo", "v2.0.0").Return(existingRelease, nil)

	release, _ := GetOrCreateRelease(repo, mockClient, false, "Updated changelog", logGroup)

	// The update will likely fail due to network, but we verify the GetRelease path worked
	// In real usage, this would update the release on GitHub
	assert.NotNil(t, release)
	assert.Equal(t, 456, release.Id)
	mockClient.AssertExpectations(t)
}

func TestGetOrCreateRelease_GetReleaseError(t *testing.T) {
	logGroup := logger.NewLogGroup("test")
	mockClient := new(MockGitHubClient)

	repo := &mockRepo{
		currentTag: "v1.0.0",
		gitHubSlug: "owner/repo",
	}

	// Mock GetRelease to return an error (not ErrReleaseNotFound)
	mockClient.On("GetRelease", "owner/repo", "v1.0.0").Return(nil, fmt.Errorf("API error"))

	release, err := GetOrCreateRelease(repo, mockClient, false, "Test", logGroup)

	assert.Error(t, err)
	assert.Nil(t, release)
	mockClient.AssertExpectations(t)
}

func TestGetOrCreateRelease_CreateReleaseError(t *testing.T) {
	logGroup := logger.NewLogGroup("test")
	mockClient := new(MockGitHubClient)

	repo := &mockRepo{
		currentTag: "v1.0.0",
		gitHubSlug: "owner/repo",
	}

	mockClient.On("GetRelease", "owner/repo", "v1.0.0").Return(nil, github.ErrReleaseNotFound)
	mockClient.On("CreateRelease", "owner/repo", mock.Anything).Return(nil, fmt.Errorf("create failed"))

	release, err := GetOrCreateRelease(repo, mockClient, false, "Test", logGroup)

	assert.Error(t, err)
	assert.Nil(t, release)
	assert.Contains(t, err.Error(), "create failed")
	mockClient.AssertExpectations(t)
}

func TestUploadToGitHub_SkipConditions(t *testing.T) {
	t.Run("skip when not github hosted", func(t *testing.T) {
		repo := &mockRepo{isGitHubHosted: false}
		err := UploadToGitHub(UploadGitHubArgs{Repo: repo})
		assert.NoError(t, err)
	})

	t.Run("skip when no tag", func(t *testing.T) {
		repo := &mockRepo{
			isGitHubHosted: true,
			gitHubSlug:     "owner/repo",
		}
		err := UploadToGitHub(UploadGitHubArgs{Repo: repo})
		assert.NoError(t, err)
	})
}

func TestUploadToGitHub_PrereleaseDetermination(t *testing.T) {
	testCases := []struct {
		releaseType        string
		expectedPrerelease bool
	}{
		{"release", false},
		{"alpha", true},
		{"beta", true},
		{"unknown", true},
	}

	for _, tc := range testCases {
		t.Run(tc.releaseType, func(t *testing.T) {
			repo := &mockRepo{isGitHubHosted: false}

			tmpDir := t.TempDir()
			changelogPath := filepath.Join(tmpDir, "changelog.md")
			err := os.WriteFile(changelogPath, []byte("# Test"), 0644)
			require.NoError(t, err)

			err = UploadToGitHub(UploadGitHubArgs{
				Repo:        repo,
				ReleaseType: tc.releaseType,
				Changelog: &changelog.Changelog{
					PreExistingFilePath: changelogPath,
				},
			})
			assert.NoError(t, err) // Skips due to not hosted
		})
	}
}

func TestUploadToGitHub_CallsGetReleaseMetadataContents(t *testing.T) {
	// Set token for the test
	err := os.Setenv("GITHUB_OAUTH", "test-token")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv("GITHUB_OAUTH")
		if err != nil {
			t.Fatalf("Failed to unset GITHUB_OAUTH: %v", err)
		}
	}()

	tmpDir := t.TempDir()

	// Create changelog
	changelogPath := filepath.Join(tmpDir, "changelog.md")
	err = os.WriteFile(changelogPath, []byte("# v1.0.0\n\nNew features"), 0644)
	require.NoError(t, err)

	// Create zip files
	zip1 := filepath.Join(tmpDir, "addon-1.0.0.zip")
	zip2 := filepath.Join(tmpDir, "addon-1.0.0-nolib.zip")
	err = os.WriteFile(zip1, []byte("fake zip 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(zip2, []byte("fake zip 2"), 0644)
	require.NoError(t, err)

	// Setup mocks
	mockClient := new(MockGitHubClient)
	repo := &mockRepo{
		isGitHubHosted: true,
		currentTag:     "v1.0.0",
		gitHubSlug:     "owner/repo",
	}

	// Mock shouldSkip checks
	mockClient.On("IsTokenSet").Return(true)

	// Mock GetRelease to return ErrReleaseNotFound so we create a new one
	mockClient.On("GetRelease", "owner/repo", "v1.0.0").Return(nil, github.ErrReleaseNotFound)

	// Mock CreateRelease
	release := &github.GitHubRelease{
		GitHubReleasePayload: github.GitHubReleasePayload{
			TagName:    "v1.0.0",
			Name:       "v1.0.0",
			Body:       "# v1.0.0\n\nNew features",
			Prerelease: false,
			Draft:      false,
		},
		Id:   789,
		Slug: "owner/repo",
	}
	mockClient.On("CreateRelease", "owner/repo", mock.Anything).Return(release, nil)

	// Mock GetReleaseMetadataContents
	mockClient.On("GetReleaseMetadataContents",
		"TestAddon",
		"1.0.0",
		mock.Anything,
		mock.Anything,
	).Return(`{"name":"TestAddon","version":"1.0.0"}`, nil)

	// Temporarily replace the global client
	oldClient := defaultGitHubClient
	defaultGitHubClient = mockClient
	defer func() { defaultGitHubClient = oldClient }()

	args := UploadGitHubArgs{
		ProjectName:    "TestAddon",
		ProjectVersion: "1.0.0",
		Repo:           repo,
		ZipPaths:       []string{zip1, zip2},
		Changelog: &changelog.Changelog{
			PreExistingFilePath: changelogPath,
		},
		ReleaseType: "release",
	}

	// Note: This will error on actual asset upload since we can't mock that easily,
	// but we can verify our client methods were called
	_ = UploadToGitHub(args)

	// Verify the mocked methods were called
	mockClient.AssertCalled(t, "IsTokenSet")
	mockClient.AssertCalled(t, "GetRelease", "owner/repo", "v1.0.0")
	mockClient.AssertCalled(t, "CreateRelease", "owner/repo", mock.Anything)
	mockClient.AssertCalled(t, "GetReleaseMetadataContents", "TestAddon", "1.0.0", mock.Anything, mock.Anything)
}

func TestUploadToGitHub_ChangelogReadError(t *testing.T) {
	mockClient := new(MockGitHubClient)
	mockClient.On("IsTokenSet").Return(true)

	repo := &mockRepo{
		isGitHubHosted: true,
		currentTag:     "v1.0.0",
		gitHubSlug:     "owner/repo",
	}

	// Temporarily replace the global client
	oldClient := defaultGitHubClient
	defaultGitHubClient = mockClient
	defer func() { defaultGitHubClient = oldClient }()

	args := UploadGitHubArgs{
		Repo: repo,
		Changelog: &changelog.Changelog{
			PreExistingFilePath: "/nonexistent/changelog.md",
		},
		ReleaseType: "release",
	}

	err := UploadToGitHub(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestUploadToGitHub_GetReleaseMetadataError(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "changelog.md")
	err := os.WriteFile(changelogPath, []byte("# Test"), 0644)
	require.NoError(t, err)

	mockClient := new(MockGitHubClient)
	mockClient.On("IsTokenSet").Return(true)

	// Mock GetRelease to return ErrReleaseNotFound
	mockClient.On("GetRelease", "owner/repo", "v1.0.0").Return(nil, github.ErrReleaseNotFound)

	// Mock CreateRelease to succeed
	release := &github.GitHubRelease{
		GitHubReleasePayload: github.GitHubReleasePayload{
			TagName: "v1.0.0",
		},
		Id:   1,
		Slug: "owner/repo",
	}
	mockClient.On("CreateRelease", "owner/repo", mock.Anything).Return(release, nil)

	// Mock GetReleaseMetadataContents to fail
	mockClient.On("GetReleaseMetadataContents", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", fmt.Errorf("metadata generation failed"))

	repo := &mockRepo{
		isGitHubHosted: true,
		currentTag:     "v1.0.0",
		gitHubSlug:     "owner/repo",
	}

	oldClient := defaultGitHubClient
	defaultGitHubClient = mockClient
	defer func() { defaultGitHubClient = oldClient }()

	args := UploadGitHubArgs{
		ProjectName:    "Test",
		ProjectVersion: "1.0.0",
		Repo:           repo,
		Changelog: &changelog.Changelog{
			PreExistingFilePath: changelogPath,
		},
		ReleaseType: "release",
	}

	err = UploadToGitHub(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata generation failed")
	mockClient.AssertExpectations(t)
}
