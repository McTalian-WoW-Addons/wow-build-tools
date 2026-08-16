package upload

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withFastRetries shrinks the retry attempt count and backoff for the
// duration of a test so exhausted-retry cases don't sit through the real
// multi-second exponential backoff.
func withFastRetries(t *testing.T, attempts int) {
	t.Helper()
	oldAttempts, oldWait := uploadMaxAttempts, uploadRetryBaseWait
	uploadMaxAttempts = attempts
	uploadRetryBaseWait = time.Millisecond
	t.Cleanup(func() {
		uploadMaxAttempts = oldAttempts
		uploadRetryBaseWait = oldWait
	})
}

func TestUploadWithRetry(t *testing.T) {
	t.Run("resends the full body on every attempt after a transient failure", func(t *testing.T) {
		withFastRetries(t, 5)

		var receivedLens []int
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			receivedLens = append(receivedLens, len(b))
			if attempts < 3 {
				// CurseForge's real failure mode: a body accompanies the
				// error status. A single *http.Request reused across
				// retries would send an empty body on this and every
				// subsequent attempt.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"errorCode":500,"errorMessage":"boom"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		payload := []byte("PAYLOAD-PAYLOAD-PAYLOAD")
		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			return http.NewRequest("POST", server.URL, bytes.NewReader(payload))
		})

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
		for i, n := range receivedLens {
			assert.Equalf(t, len(payload), n, "attempt %d sent %d bytes, want the full %d-byte payload", i+1, n, len(payload))
		}
	})

	t.Run("calls newReq once per attempt instead of reusing one request", func(t *testing.T) {
		// This is the actual contract that prevents the CurseForge bug: a
		// *http.Request's body reader is drained by client.Do, so reusing
		// one across retries silently sends an empty body on attempt 2+.
		// Asserting the builder call count catches a regression to that
		// pattern even though, over a loopback connection, Go's Transport
		// sometimes self-heals a reused, drained body via req.GetBody and
		// hides the symptom on the wire.
		withFastRetries(t, 4)

		attempts, builderCalls := 0, 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			builderCalls++
			return http.NewRequest("POST", server.URL, bytes.NewReader([]byte("x")))
		})

		require.Error(t, err)
		assert.Equal(t, uploadMaxAttempts, attempts)
		assert.Equal(t, attempts, builderCalls, "newReq must be called once per attempt")
	})

	t.Run("stops retrying on a 400 response", func(t *testing.T) {
		withFastRetries(t, 5)

		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`<html>bad request</html>`))
		}))
		defer server.Close()

		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			return http.NewRequest("POST", server.URL, bytes.NewReader([]byte("x")))
		})

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("stops retrying on a 422 response", func(t *testing.T) {
		withFastRetries(t, 5)

		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusUnprocessableEntity)
		}))
		defer server.Close()

		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			return http.NewRequest("POST", server.URL, bytes.NewReader([]byte("x")))
		})

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("gives up after exhausting all attempts on repeated 500s", func(t *testing.T) {
		withFastRetries(t, 3)

		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			return http.NewRequest("POST", server.URL, bytes.NewReader([]byte("x")))
		})

		require.Error(t, err)
		assert.Equal(t, uploadMaxAttempts, attempts)
	})

	t.Run("returns an error if the request builder fails", func(t *testing.T) {
		withFastRetries(t, 5)

		err := uploadWithRetry(logger.NewLogGroup("test"), "done", func() (*http.Request, error) {
			return nil, assert.AnError
		})

		require.Error(t, err)
	})
}
