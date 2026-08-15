package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTestServer points githubApiUrl at a local httptest.Server for the
// duration of the test and restores it afterward.
func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	origApiUrl := githubApiUrl
	githubApiUrl = server.URL + "/"
	t.Cleanup(func() { githubApiUrl = origApiUrl })

	origAuthHeaderValue := authHeaderValue
	authHeaderValue = ""
	t.Cleanup(func() { authHeaderValue = origAuthHeaderValue })
	t.Setenv("GITHUB_OAUTH", "test-token")
}

func TestCreateRelease(t *testing.T) {
	t.Run("successful creation returns the decoded release", func(t *testing.T) {
		withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/owner/repo/releases", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(GitHubReleasePayload{
				TagName: "v1.0.0",
				Name:    "v1.0.0",
				Body:    "release notes",
			})
		})

		release, err := CreateRelease("owner/repo", GitHubReleasePayload{
			TagName: "v1.0.0",
			Name:    "v1.0.0",
			Body:    "release notes",
		})

		require.NoError(t, err)
		require.NotNil(t, release)
		assert.Equal(t, "v1.0.0", release.TagName)
		assert.Equal(t, "release notes", release.Body)
		assert.Equal(t, "owner/repo", release.Slug)
	})

	t.Run("non-201 response returns an error", func(t *testing.T) {
		withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		})

		release, err := CreateRelease("owner/repo", GitHubReleasePayload{TagName: "v1.0.0"})

		assert.Error(t, err)
		assert.Nil(t, release)
	})
}
