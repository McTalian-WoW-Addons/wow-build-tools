package github

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTokenSet(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "token is set",
			token:    "test-token",
			expected: true,
		},
		{
			name:     "token is empty",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original
			origToken := os.Getenv("GITHUB_OAUTH")

			if tt.token != "" {
				_ = os.Setenv("GITHUB_OAUTH", tt.token)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}

			result := IsTokenSet()
			assert.Equal(t, tt.expected, result)

			// Restore
			if origToken != "" {
				_ = os.Setenv("GITHUB_OAUTH", origToken)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}
		})
	}
}

func TestGetAuthHeaderValue(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
		expected    string
	}{
		{
			name:        "with valid token",
			token:       "test-token-123",
			expectError: false,
			expected:    "token test-token-123",
		},
		{
			name:        "without token",
			token:       "",
			expectError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and reset
			origToken := os.Getenv("GITHUB_OAUTH")
			origAuthHeaderValue := authHeaderValue
			authHeaderValue = "" // Reset cached value

			if tt.token != "" {
				_ = os.Setenv("GITHUB_OAUTH", tt.token)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}

			result, err := getAuthHeaderValue()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			// Restore
			if origToken != "" {
				_ = os.Setenv("GITHUB_OAUTH", origToken)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}
			authHeaderValue = origAuthHeaderValue
		})
	}
}

func TestGetAuthHeaderValue_Caching(t *testing.T) {
	// Save and reset
	origToken := os.Getenv("GITHUB_OAUTH")
	origAuthHeaderValue := authHeaderValue
	authHeaderValue = ""

	_ = os.Setenv("GITHUB_OAUTH", "test-token")

	// First call
	result1, err := getAuthHeaderValue()
	require.NoError(t, err)
	assert.Equal(t, "token test-token", result1)

	// Change the token
	_ = os.Setenv("GITHUB_OAUTH", "different-token")

	// Second call should return cached value
	result2, err := getAuthHeaderValue()
	require.NoError(t, err)
	assert.Equal(t, "token test-token", result2, "should return cached value")

	// Restore
	if origToken != "" {
		_ = os.Setenv("GITHUB_OAUTH", origToken)
	} else {
		_ = os.Unsetenv("GITHUB_OAUTH")
	}
	authHeaderValue = origAuthHeaderValue
}

func TestAddAcceptHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	addAcceptHeader(req)

	assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))
}

func TestAddAuthHeader(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "with valid token",
			token:       "test-token",
			expectError: false,
		},
		{
			name:        "without token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and reset
			origToken := os.Getenv("GITHUB_OAUTH")
			origAuthHeaderValue := authHeaderValue
			authHeaderValue = ""

			if tt.token != "" {
				_ = os.Setenv("GITHUB_OAUTH", tt.token)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}

			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(t, err)

			err = addAuthHeader(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "token "+tt.token, req.Header.Get("Authorization"))
			}

			// Restore
			if origToken != "" {
				_ = os.Setenv("GITHUB_OAUTH", origToken)
			} else {
				_ = os.Unsetenv("GITHUB_OAUTH")
			}
			authHeaderValue = origAuthHeaderValue
		})
	}
}
