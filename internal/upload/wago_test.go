package upload

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocateWagoId(t *testing.T) {
	t.Run("single wago id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WagoId: "abc123"},
			{WagoId: "abc123"},
		}
		wagoId, err := locateWagoId(tocFiles)
		assert.NoError(t, err)
		assert.Equal(t, "abc123", wagoId)
	})

	t.Run("no wago id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WagoId: ""},
			{WagoId: ""},
		}
		_, err := locateWagoId(tocFiles)
		assert.ErrorIs(t, err, ErrNoWagoId)
	})

	t.Run("multiple different wago ids", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WagoId: "abc123"},
			{WagoId: "xyz789"},
		}
		_, err := locateWagoId(tocFiles)
		assert.ErrorIs(t, err, ErrMultipleWagoIds)
	})
}

func TestGetWagoId(t *testing.T) {
	t.Run("flag overrides toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WagoId: "abc123"}}
		wagoId, err := getWagoId(tocFiles, "xyz789")
		assert.NoError(t, err)
		assert.Equal(t, "xyz789", wagoId)
	})

	t.Run("zero flag clears id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WagoId: "abc123"}}
		_, err := getWagoId(tocFiles, "0")
		assert.ErrorIs(t, err, ErrNoWagoId)
	})

	t.Run("empty flag uses toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WagoId: "abc123"}}
		wagoId, err := getWagoId(tocFiles, "")
		assert.NoError(t, err)
		assert.Equal(t, "abc123", wagoId)
	})
}

func TestWagoUploadLookupToken(t *testing.T) {
	t.Run("token found", func(t *testing.T) {
		err := os.Setenv("WAGO_API_TOKEN", "test-wago-token")
		assert.NoError(t, err)
		defer func() {
			err := os.Unsetenv("WAGO_API_TOKEN")
			if err != nil {
				t.Fatalf("Failed to unset WAGO_API_TOKEN: %v", err)
			}
		}()

		wu := &wagoUpload{}
		err = wu.lookupWagoToken()
		assert.NoError(t, err)
		assert.Equal(t, "test-wago-token", wu.token)
	})

	t.Run("token not found", func(t *testing.T) {
		err := os.Unsetenv("WAGO_API_TOKEN")
		if err != nil {
			t.Fatalf("Failed to unset WAGO_API_TOKEN: %v", err)
		}
		wu := &wagoUpload{}
		err = wu.lookupWagoToken()
		assert.ErrorIs(t, err, ErrNoWagoApiKey)
	})
}

func TestWagoPayloadMarshalJSON(t *testing.T) {
	payload := wagoPayload{
		Label:     "Test Release",
		Stability: "stable",
		Changelog: "Test changelog",
		SupportedPatches: map[string][]string{
			"retail":  {"11.0.5", "11.0.7"},
			"classic": {"1.15.5"},
		},
	}

	data, err := json.Marshal(&payload)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "Test Release", result["label"])
	assert.Equal(t, "stable", result["stability"])
	assert.Equal(t, "Test changelog", result["changelog"])
	assert.NotNil(t, result["supported_retail_patches"])
	assert.NotNil(t, result["supported_classic_patches"])
}

func TestWagoUploadPreparePayload(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "changelog.md")
	err := os.WriteFile(changelogPath, []byte("# Changelog\n\nTest changes"), 0644)
	require.NoError(t, err)

	wu := &wagoUpload{
		displayName:    "v1.0.0",
		stabilityValue: "stable",
		changelog: &changelog.Changelog{
			PreExistingFilePath: changelogPath,
		},
		supportMap: map[string][]string{
			"retail": {"11.0.5"},
		},
	}

	err = wu.preparePayload()
	assert.NoError(t, err)
	assert.NotEmpty(t, wu.metadataPart)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(wu.metadataPart), &result)
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", result["label"])
	assert.Equal(t, "stable", result["stability"])
	assert.Contains(t, result["changelog"], "Test changes")
}

func TestWagoUploadUpload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-token")

			err := r.ParseMultipartForm(10 << 20)
			assert.NoError(t, err)
			assert.NotEmpty(t, r.FormValue("metadata"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip"), 0644)
		require.NoError(t, err)

		wu := &wagoUpload{
			token:        "test-token",
			uploadUrl:    server.URL,
			zipFile:      zipPath,
			metadataPart: `{"label":"test","stability":"stable","changelog":"test"}`,
			logGroup:     logger.NewLogGroup("test"),
		}

		err = wu.upload()
		assert.NoError(t, err)
	})

	t.Run("upload with retry", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip"), 0644)
		require.NoError(t, err)

		wu := &wagoUpload{
			token:        "test-token",
			uploadUrl:    server.URL,
			zipFile:      zipPath,
			metadataPart: `{"label":"test","stability":"stable","changelog":"test"}`,
			logGroup:     logger.NewLogGroup("test"),
		}

		err = wu.upload()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, attempts, 2)
	})
}

func TestUploadToWago(t *testing.T) {
	t.Run("skip when skip upload is true", func(t *testing.T) {
		err := UploadToWago(UploadWagoArgs{
			SkipUpload: true,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no wago id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WagoId: ""}}
		err := UploadToWago(UploadWagoArgs{
			TocFiles: tocFiles,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no token", func(t *testing.T) {
		err := os.Unsetenv("WAGO_API_TOKEN")
		if err != nil {
			t.Fatalf("Failed to unset WAGO_API_TOKEN: %v", err)
		}
		tocFiles := []*toc.Toc{{WagoId: "abc123"}}
		err = UploadToWago(UploadWagoArgs{
			TocFiles: tocFiles,
			WagoId:   "abc123",
		})
		assert.NoError(t, err)
	})

	t.Run("map release type correctly", func(t *testing.T) {
		err := os.Unsetenv("WAGO_API_TOKEN")
		if err != nil {
			t.Fatalf("Failed to unset WAGO_API_TOKEN: %v", err)
		}

		testCases := []struct {
			input    string
			expected string
		}{
			{"alpha", "alpha"},
			{"beta", "beta"},
			{"release", "stable"},
			{"unknown", "alpha"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				tocFiles := []*toc.Toc{{WagoId: "abc123"}}
				err := UploadToWago(UploadWagoArgs{
					TocFiles:    tocFiles,
					ReleaseType: tc.input,
				})
				assert.NoError(t, err)
			})
		}
	})
}
