package toc

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useWagoServer points the wago client at srv with a short timeout and resets
// the build cache, restoring everything on cleanup.
func useWagoServer(t *testing.T, srv *httptest.Server, timeout time.Duration) {
	t.Helper()
	oldURL, oldTimeout, oldBackoff, oldCache := wagoApiUrl, wagoTimeout, wagoRetryBackoff, cacheLatestBuilds
	wagoApiUrl, wagoTimeout, wagoRetryBackoff, cacheLatestBuilds = srv.URL, timeout, time.Millisecond, nil
	t.Cleanup(func() {
		wagoApiUrl, wagoTimeout, wagoRetryBackoff, cacheLatestBuilds = oldURL, oldTimeout, oldBackoff, oldCache
	})
}

func TestGetLatestBuildInfo_Timeout(t *testing.T) {
	var hits atomic.Int32
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(done) })
	useWagoServer(t, srv, 50*time.Millisecond)

	builds, err := GetLatestBuildInfo()
	assert.Error(t, err)
	assert.Nil(t, builds)
	assert.EqualValues(t, 2, hits.Load(), "timeout should be retried once")
}

func TestGetLatestBuildInfo_504(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	t.Cleanup(srv.Close)
	useWagoServer(t, srv, time.Second)

	_, err := GetLatestBuildInfo()
	assert.ErrorContains(t, err, "unexpected status code: 504")
	assert.EqualValues(t, 2, hits.Load(), "5xx should be retried once")
}

func TestGetLatestBuildInfo_4xxNotRetried(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	useWagoServer(t, srv, time.Second)

	_, err := GetLatestBuildInfo()
	assert.Error(t, err)
	assert.EqualValues(t, 1, hits.Load())
}

func TestGetLatestBuildInfo_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/builds/latest", r.URL.Path)
		_, _ = w.Write([]byte(`{"wow":{"product":"wow","version":"12.0.1.12345"}}`))
	}))
	t.Cleanup(srv.Close)
	useWagoServer(t, srv, time.Second)

	builds, err := GetLatestBuildInfo()
	require.NoError(t, err)
	assert.Equal(t, "12.0.1.12345", (*builds)[ProductWow].Version)
}

func writeAddon(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "MyAddon")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MyAddon.toc"), []byte("## Interface: 110000\n## Title: MyAddon\n\nMyAddon.lua\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MyAddon.lua"), []byte("-- x\n"), 0o644))
	return dir
}

func TestRunTocCheck_DegradesWhenWagoFails(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"timeout": func(w http.ResponseWriter, r *http.Request) { time.Sleep(300 * time.Millisecond) },
		"504":     func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusGatewayTimeout) },
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(h)
			t.Cleanup(srv.Close)
			useWagoServer(t, srv, 50*time.Millisecond)

			oldDir, oldCheck := TocParams.AddonDir, TocCheckParams
			TocParams.AddonDir = writeAddon(t)
			TocCheckParams.SkipInterfaceCheck = false
			t.Cleanup(func() { TocParams.AddonDir, TocCheckParams = oldDir, oldCheck })

			assert.NoError(t, RunTocCheck())
		})
	}
}

func TestRunTocCheck_RealFailureStillFailsWhenWagoDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	t.Cleanup(srv.Close)
	useWagoServer(t, srv, time.Second)

	oldDir, oldCheck := TocParams.AddonDir, TocCheckParams
	dir := writeAddon(t)
	// Rename the folder so the name check fails.
	bad := filepath.Join(filepath.Dir(dir), "WrongName")
	require.NoError(t, os.Rename(dir, bad))
	TocParams.AddonDir = bad
	TocCheckParams.SkipInterfaceCheck = false
	t.Cleanup(func() { TocParams.AddonDir, TocCheckParams = oldDir, oldCheck })

	assert.Error(t, RunTocCheck())
}
