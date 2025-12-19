package repo

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

// Test simple getter methods with 0% coverage

func TestGitRepo_GetGitHubSlug(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		expected string
	}{
		{
			name:     "valid slug",
			slug:     "owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "empty slug",
			slug:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				gitHubSlug: tt.slug,
			}
			assert.Equal(t, tt.expected, gR.GetGitHubSlug())
		})
	}
}

func TestGitRepo_GetGitHubOwnerAndRepo(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "valid slug with owner and repo",
			slug:          "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "empty slug",
			slug:          "",
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			name:          "slug with only owner",
			slug:          "owner",
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			name:          "slug with multiple segments",
			slug:          "owner/repo/extra",
			expectedOwner: "",
			expectedRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				gitHubSlug: tt.slug,
			}
			owner, repo := gR.GetGitHubOwnerAndRepo()
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
		})
	}
}

func TestGitRepo_IsGitHubHosted(t *testing.T) {
	tests := []struct {
		name     string
		isGitHub bool
		expected bool
	}{
		{
			name:     "github hosted",
			isGitHub: true,
			expected: true,
		},
		{
			name:     "not github hosted",
			isGitHub: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				isGitHubUrl: tt.isGitHub,
			}
			assert.Equal(t, tt.expected, gR.IsGitHubHosted())
		})
	}
}

func TestGitRepo_GetCurrentTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{
			name:     "with tag",
			tag:      "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				BaseVcsRepo: BaseVcsRepo{
					CurrentTag: tt.tag,
				},
			}
			assert.Equal(t, tt.expected, gR.GetCurrentTag())
		})
	}
}

func TestGitRepo_GetPreviousVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "with version",
			version:  "v0.9.0",
			expected: "v0.9.0",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				BaseVcsRepo: BaseVcsRepo{
					PreviousVersion: tt.version,
				},
			}
			assert.Equal(t, tt.expected, gR.GetPreviousVersion())
		})
	}
}

func TestGitRepo_GetProjectVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "with version",
			version:  "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gR := &GitRepo{
				BaseVcsRepo: BaseVcsRepo{
					ProjectVersion: tt.version,
				},
			}
			assert.Equal(t, tt.expected, gR.GetProjectVersion())
		})
	}
}

func TestGitRepo_GetRepoRoot(t *testing.T) {
	mockRepo := &Repo{
		repoRoot: "/test/path",
	}

	gR := &GitRepo{
		repo: mockRepo,
	}

	assert.Equal(t, "/test/path", gR.GetRepoRoot())
}

func TestNewTagNameHash(t *testing.T) {
	tests := []struct {
		name         string
		tagName      string
		hash         string
		expectedName string
		expectedHash string
	}{
		{
			name:         "valid tag and hash",
			tagName:      "v1.0.0",
			hash:         "abc123def456",
			expectedName: "v1.0.0",
			expectedHash: "abc123def456",
		},
		{
			name:         "empty tag name",
			tagName:      "",
			hash:         "abc123def456",
			expectedName: "",
			expectedHash: "abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := plumbing.NewHash(tt.hash)
			result := newTagNameHash(tt.tagName, hash)
			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, hash, result.Hash)
		})
	}
}

func TestSortTagNameHashes(t *testing.T) {
	tests := []struct {
		name     string
		input    []tagNameHash
		expected []string
	}{
		{
			name: "sort semantic versions descending",
			input: []tagNameHash{
				{Name: "v1.0.0", Hash: plumbing.NewHash("hash1")},
				{Name: "v2.0.0", Hash: plumbing.NewHash("hash2")},
				{Name: "v1.5.0", Hash: plumbing.NewHash("hash3")},
			},
			expected: []string{"v2.0.0", "v1.5.0", "v1.0.0"},
		},
		{
			name: "sort versions without v prefix",
			input: []tagNameHash{
				{Name: "1.0.0", Hash: plumbing.NewHash("hash1")},
				{Name: "2.0.0", Hash: plumbing.NewHash("hash2")},
				{Name: "1.5.0", Hash: plumbing.NewHash("hash3")},
			},
			expected: []string{"2.0.0", "1.5.0", "1.0.0"},
		},
		{
			name: "sort patch versions",
			input: []tagNameHash{
				{Name: "v1.0.1", Hash: plumbing.NewHash("hash1")},
				{Name: "v1.0.3", Hash: plumbing.NewHash("hash2")},
				{Name: "v1.0.2", Hash: plumbing.NewHash("hash3")},
			},
			expected: []string{"v1.0.3", "v1.0.2", "v1.0.1"},
		},
		{
			name: "sort major.minor versions",
			input: []tagNameHash{
				{Name: "v1.1", Hash: plumbing.NewHash("hash1")},
				{Name: "v2.0", Hash: plumbing.NewHash("hash2")},
				{Name: "v1.9", Hash: plumbing.NewHash("hash3")},
			},
			expected: []string{"v2.0", "v1.9", "v1.1"},
		},
		{
			name: "single version",
			input: []tagNameHash{
				{Name: "v1.0.0", Hash: plumbing.NewHash("hash1")},
			},
			expected: []string{"v1.0.0"},
		},
		{
			name:     "empty slice",
			input:    []tagNameHash{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortTagNameHashes(tt.input)
			result := make([]string, len(tt.input))
			for i, tag := range tt.input {
				result[i] = tag.Name
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortTagNameHashes_MultiDigitVersions(t *testing.T) {
	input := []tagNameHash{
		{Name: "v1.9.0", Hash: plumbing.NewHash("hash1")},
		{Name: "v1.10.0", Hash: plumbing.NewHash("hash2")},
		{Name: "v1.11.0", Hash: plumbing.NewHash("hash3")},
		{Name: "v2.0.0", Hash: plumbing.NewHash("hash4")},
	}

	sortTagNameHashes(input)

	expected := []string{"v2.0.0", "v1.11.0", "v1.10.0", "v1.9.0"}
	result := make([]string, len(input))
	for i, tag := range input {
		result[i] = tag.Name
	}

	assert.Equal(t, expected, result)
}
