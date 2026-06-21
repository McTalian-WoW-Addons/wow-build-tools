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

func TestLocateWowiId(t *testing.T) {
	t.Run("single wowi id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WowiId: "54321"},
			{WowiId: "54321"},
		}
		wowiId, err := locateWowiId(tocFiles)
		assert.NoError(t, err)
		assert.Equal(t, "54321", wowiId)
	})

	t.Run("no wowi id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WowiId: ""},
			{WowiId: ""},
		}
		_, err := locateWowiId(tocFiles)
		assert.ErrorIs(t, err, ErrNoWowiId)
	})

	t.Run("multiple different wowi ids", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{WowiId: "54321"},
			{WowiId: "98765"},
		}
		_, err := locateWowiId(tocFiles)
		assert.ErrorIs(t, err, ErrMultipleWowiIds)
	})
}

func TestGetWowiId(t *testing.T) {
	t.Run("flag overrides toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WowiId: "54321"}}
		wowiId, err := getWowiId(tocFiles, "99999")
		assert.NoError(t, err)
		assert.Equal(t, "99999", wowiId)
	})

	t.Run("zero flag clears id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WowiId: "54321"}}
		_, err := getWowiId(tocFiles, "0")
		assert.ErrorIs(t, err, ErrNoWowiId)
	})

	t.Run("empty flag uses toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WowiId: "54321"}}
		wowiId, err := getWowiId(tocFiles, "")
		assert.NoError(t, err)
		assert.Equal(t, "54321", wowiId)
	})
}

func TestWowiUploadLookupToken(t *testing.T) {
	t.Run("token found", func(t *testing.T) {
		err := os.Setenv("WOWI_API_TOKEN", "test-wowi-token")
		assert.NoError(t, err)
		defer func() {
			err := os.Unsetenv("WOWI_API_TOKEN")
			if err != nil {
				t.Fatalf("Failed to unset WOWI_API_TOKEN: %v", err)
			}
		}()

		wu := &wowiUpload{}
		err = wu.lookupWowiToken()
		assert.NoError(t, err)
		assert.Equal(t, "test-wowi-token", wu.token)
	})

	t.Run("token not found", func(t *testing.T) {
		err := os.Unsetenv("WOWI_API_TOKEN")
		if err != nil {
			t.Fatalf("Failed to unset WOWI_API_TOKEN: %v", err)
		}
		wu := &wowiUpload{}
		err = wu.lookupWowiToken()
		assert.ErrorIs(t, err, ErrNoWowiApiKey)
	})
}

func TestStringInSlice(t *testing.T) {
	t.Run("string found", func(t *testing.T) {
		list := []string{"a", "b", "c"}
		assert.True(t, stringInSlice("b", list))
	})

	t.Run("string not found", func(t *testing.T) {
		list := []string{"a", "b", "c"}
		assert.False(t, stringInSlice("d", list))
	})

	t.Run("empty list", func(t *testing.T) {
		list := []string{}
		assert.False(t, stringInSlice("a", list))
	})
}

func TestWowiUploadValidateGameVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		versions := []wowiGameVersionsEntry{
			{Id: "11.0.5", Name: "11.0.5", Game: "retail", Interface: "110005"},
			{Id: "1.15.5", Name: "1.15.5", Game: "classic", Interface: "11505"},
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(versions)
		assert.NoError(t, err)
	}))
	defer server.Close()

	oldUrl := wowiGameVersionsUrl
	wowiGameVersionsUrl = server.URL
	defer func() { wowiGameVersionsUrl = oldUrl }()

	wu := &wowiUpload{
		logGroup: logger.NewLogGroup("test"),
	}

	err := wu.validateGameVersions([]string{"11.0.5", "1.15.5"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"11.0.5", "1.15.5"}, wu.compatible)
}

func TestWowiUploadUpload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
			assert.Equal(t, "test-token", r.Header.Get("x-api-token"))

			err := r.ParseMultipartForm(10 << 20)
			assert.NoError(t, err)
			assert.Equal(t, "12345", r.FormValue("id"))
			assert.Equal(t, "1.0.0", r.FormValue("version"))
			assert.NotEmpty(t, r.FormValue("compatible"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		oldUrl := wowiUploadUrl
		wowiUploadUrl = server.URL
		defer func() { wowiUploadUrl = oldUrl }()

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip"), 0644)
		require.NoError(t, err)

		changelogPath := filepath.Join(tmpDir, "changelog.txt")
		err = os.WriteFile(changelogPath, []byte("Test changelog"), 0644)
		require.NoError(t, err)

		wu := &wowiUpload{
			token:      "test-token",
			projectId:  "12345",
			zipFile:    zipPath,
			version:    "1.0.0",
			compatible: []string{"11.0.5"},
			changelog: &changelog.Changelog{
				PreExistingFilePath: changelogPath,
			},
			archiveOld: false,
			logGroup:   logger.NewLogGroup("test"),
		}

		err = wu.upload()
		assert.NoError(t, err)
	})

	t.Run("upload with retry", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		oldUrl := wowiUploadUrl
		wowiUploadUrl = server.URL
		defer func() { wowiUploadUrl = oldUrl }()

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip"), 0644)
		require.NoError(t, err)

		wu := &wowiUpload{
			token:      "test-token",
			projectId:  "12345",
			zipFile:    zipPath,
			version:    "1.0.0",
			compatible: []string{"11.0.5"},
			archiveOld: false,
			logGroup:   logger.NewLogGroup("test"),
		}

		err = wu.upload()
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})
}

func TestUploadToWowi(t *testing.T) {
	t.Run("skip when skip upload is true", func(t *testing.T) {
		err := UploadToWowi(UploadWowiArgs{
			SkipUpload: true,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no wowi id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{WowiId: ""}}
		err := UploadToWowi(UploadWowiArgs{
			TocFiles: tocFiles,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no token", func(t *testing.T) {
		err := os.Unsetenv("WOWI_API_TOKEN")
		if err != nil {
			t.Fatalf("Failed to unset WOWI_API_TOKEN: %v", err)
		}
		tocFiles := []*toc.Toc{{WowiId: "54321"}}
		err = UploadToWowi(UploadWowiArgs{
			TocFiles: tocFiles,
			WowiId:   "54321",
		})
		assert.NoError(t, err)
	})
}
