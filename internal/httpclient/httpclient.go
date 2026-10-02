// Package httpclient provides HTTP clients with bounded timeouts so that a
// slow or unreachable upstream cannot hang the CLI indefinitely.
package httpclient

import (
	"net/http"
	"time"
)

// LookupTimeout bounds small metadata lookups (build info, game versions).
const LookupTimeout = 10 * time.Second

// NewLookup returns a client suitable for small GET lookups.
func NewLookup() *http.Client {
	return &http.Client{Timeout: LookupTimeout}
}
