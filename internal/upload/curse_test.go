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
	"github.com/McTalian/wow-build-tools/internal/pkg"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocateCurseId(t *testing.T) {
	t.Run("single curse id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{CurseId: "12345"},
			{CurseId: "12345"},
		}
		curseId, err := locateCurseId(tocFiles)
		assert.NoError(t, err)
		assert.Equal(t, "12345", curseId)
	})

	t.Run("no curse id", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{CurseId: ""},
			{CurseId: ""},
		}
		_, err := locateCurseId(tocFiles)
		assert.ErrorIs(t, err, ErrNoCurseId)
	})

	t.Run("multiple different curse ids", func(t *testing.T) {
		tocFiles := []*toc.Toc{
			{CurseId: "12345"},
			{CurseId: "67890"},
		}
		_, err := locateCurseId(tocFiles)
		assert.ErrorIs(t, err, ErrMultipleCurseIds)
	})
}

func TestGetCurseId(t *testing.T) {
	t.Run("flag overrides toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{CurseId: "12345"}}
		curseId, err := getCurseId(tocFiles, "99999")
		assert.NoError(t, err)
		assert.Equal(t, "99999", curseId)
	})

	t.Run("zero flag clears id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{CurseId: "12345"}}
		_, err := getCurseId(tocFiles, "0")
		assert.ErrorIs(t, err, ErrNoCurseId)
	})

	t.Run("empty flag uses toc", func(t *testing.T) {
		tocFiles := []*toc.Toc{{CurseId: "12345"}}
		curseId, err := getCurseId(tocFiles, "")
		assert.NoError(t, err)
		assert.Equal(t, "12345", curseId)
	})
}

func TestCurseUploadLookupToken(t *testing.T) {
	t.Run("token found", func(t *testing.T) {
		err := os.Setenv("CF_API_KEY", "test-token")
		assert.NoError(t, err)
		defer func() {
			err := os.Unsetenv("CF_API_KEY")
			if err != nil {
				t.Fatalf("Failed to unset CF_API_KEY: %v", err)
			}
		}()

		cu := &curseUpload{}
		err = cu.lookupCurseToken()
		assert.NoError(t, err)
		assert.Equal(t, "test-token", cu.token)
	})

	t.Run("token not found", func(t *testing.T) {
		err := os.Unsetenv("CF_API_KEY")
		if err != nil {
			t.Fatalf("Failed to unset CF_API_KEY: %v", err)
		}
		cu := &curseUpload{}
		err = cu.lookupCurseToken()
		assert.ErrorIs(t, err, ErrNoCurseApiKey)
	})
}

func TestCurseUploadValidateGameVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "test-token", r.Header.Get("x-api-token"))

		versions := curseGameVersionResponse{
			{ID: 1, Name: "11.0.5", GameVersionTypeID: 517},
			{ID: 2, Name: "1.15.5", GameVersionTypeID: 67408},
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(versions)
		assert.NoError(t, err)
	}))
	defer server.Close()

	oldUrl := curseGameVersionsUrl
	curseGameVersionsUrl = server.URL
	defer func() { curseGameVersionsUrl = oldUrl }()

	cu := &curseUpload{
		token:    "test-token",
		logGroup: logger.NewLogGroup("test"),
	}

	err := cu.validateGameVersions([]string{"11.0.5", "1.15.5"})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, cu.gameVersions)
}

func TestCurseUploadPreparePayload(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "changelog.txt")
	err := os.WriteFile(changelogPath, []byte("Test changelog"), 0644)
	require.NoError(t, err)

	cu := &curseUpload{
		displayName:  "Test Release",
		releaseType:  BetaRelease,
		gameVersions: []int{1, 2},
		changelog: &changelog.Changelog{
			PreExistingFilePath: changelogPath,
			MarkupType:          changelog.TextMT,
		},
	}

	pkgMeta := &pkg.PkgMeta{
		EmbeddedLibraries:    []string{"LibStub"},
		ToolsUsed:            []string{"wow-build-tools"},
		RequiredDependencies: []string{"BigWigs"},
		OptionalDependencies: []string{"WeakAuras"},
	}

	err = cu.preparePayload(pkgMeta)
	assert.NoError(t, err)
	assert.NotEmpty(t, cu.metadataPart)

	var payload cursePayload
	err = json.Unmarshal([]byte(cu.metadataPart), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "Test changelog", payload.Changelog)
	assert.Equal(t, "Test Release", payload.DisplayName)
	assert.Equal(t, BetaRelease, payload.ReleaseType)
}

func TestCurseUploadUpload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
			assert.Equal(t, "test-token", r.Header.Get("x-api-token"))

			err := r.ParseMultipartForm(10 << 20)
			assert.NoError(t, err)
			assert.NotEmpty(t, r.FormValue("metadata"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip content"), 0644)
		require.NoError(t, err)

		cu := &curseUpload{
			token:        "test-token",
			uploadUrl:    server.URL,
			zipFile:      zipPath,
			metadataPart: `{"changelog":"test","changelogType":"text","displayName":"test","releaseType":"alpha","gameVersions":[1]}`,
			logGroup:     logger.NewLogGroup("test"),
		}

		err = cu.upload()
		assert.NoError(t, err)
	})

	t.Run("upload retry on failure", func(t *testing.T) {
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

		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(zipPath, []byte("fake zip content"), 0644)
		require.NoError(t, err)

		cu := &curseUpload{
			token:        "test-token",
			uploadUrl:    server.URL,
			zipFile:      zipPath,
			metadataPart: `{"changelog":"test","changelogType":"text","displayName":"test","releaseType":"alpha","gameVersions":[1]}`,
			logGroup:     logger.NewLogGroup("test"),
		}

		err = cu.upload()
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})
}

func TestUploadToCurse(t *testing.T) {
	t.Run("skip when skip upload is true", func(t *testing.T) {
		err := UploadToCurse(UploadCurseArgs{
			SkipUpload: true,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when only localization is true", func(t *testing.T) {
		err := UploadToCurse(UploadCurseArgs{
			OnlyLocalization: true,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no curse id", func(t *testing.T) {
		tocFiles := []*toc.Toc{{CurseId: ""}}
		err := UploadToCurse(UploadCurseArgs{
			TocFiles: tocFiles,
		})
		assert.NoError(t, err)
	})

	t.Run("skip when no token", func(t *testing.T) {
		err := os.Unsetenv("CF_API_KEY")
		if err != nil {
			t.Fatalf("Failed to unset CF_API_KEY: %v", err)
		}
		tocFiles := []*toc.Toc{{CurseId: "12345"}}
		err = UploadToCurse(UploadCurseArgs{
			TocFiles: tocFiles,
			CurseId:  "12345",
		})
		assert.NoError(t, err)
	})
}
