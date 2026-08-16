package upload

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/McTalian/wow-build-tools/internal/changelog"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
)

// var, not const, so tests can shrink attempts/backoff to keep the suite fast.
var (
	uploadMaxAttempts   = 5
	uploadRetryBaseWait = 2 * time.Second
)

// uploadWithRetry sends an HTTP request built by newReq, retrying transient
// failures with exponential backoff. newReq is called once per attempt so
// each attempt gets a fresh, unconsumed request body -- reusing a single
// *http.Request across retries silently sends a truncated body on any
// attempt after the first, since client.Do drains the body reader.
//
// 400 and 422 responses are treated as terminal (request-shape errors that
// won't succeed on retry) rather than transient.
func uploadWithRetry(logGroup *logger.LogGroup, successMsg string, newReq func() (*http.Request, error)) error {
	client := &http.Client{}
	delay := uploadRetryBaseWait

	var lastErr error
	for attempt := 1; attempt <= uploadMaxAttempts; attempt++ {
		req, err := newReq()
		if err != nil {
			return fmt.Errorf("failed to build request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			logGroup.Info("%s", successMsg)
			_ = resp.Body.Close()
			return nil
		}

		if err != nil {
			logGroup.Warn("upload error: %v", err)
			lastErr = err
		} else {
			logGroup.Warn("unexpected status code: %d", resp.StatusCode)
			jsonBody := map[string]interface{}{}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&jsonBody); decodeErr != nil {
				logGroup.Warn("failed to decode response body: %v", decodeErr)
			} else {
				logGroup.Warn("response body: %v", jsonBody)
			}
			status, statusCode := resp.Status, resp.StatusCode
			_ = resp.Body.Close()

			if statusCode == http.StatusUnprocessableEntity || statusCode == http.StatusBadRequest {
				return fmt.Errorf("upload failed: %s", status)
			}
			lastErr = fmt.Errorf("upload failed with status code %d", statusCode)
		}

		if attempt < uploadMaxAttempts {
			logGroup.Warn("Retrying: Attempt %d/%d in %s...", attempt+1, uploadMaxAttempts, delay)
			time.Sleep(delay)
			delay *= 2
		}
	}

	return fmt.Errorf("upload failed after %d attempts: %w", uploadMaxAttempts, lastErr)
}

type UploadArgs struct {
	Input             string
	Label             string
	InterfaceVersions []int
	Changelog         string
	ReleaseType       string
}

var UploadParams = &UploadArgs{}

type preparePayload struct {
	toc       *toc.Toc
	changelog *changelog.Changelog
	cleanup   func() // cleanup function to remove temp files
}

func prepareUpload() (prepPayload *preparePayload, err error) {
	defer func() {
		if err != nil {
			logger.Error("Failed to prepare upload: %v", err)
		}
	}()

	var tempFiles []string // Track temp files for cleanup

	tmp := os.TempDir()
	tmpToc, err := os.CreateTemp(tmp, "wbt*.toc")
	if err != nil {
		return
	}
	tempFiles = append(tempFiles, tmpToc.Name())
	defer func() {
		_ = tmpToc.Close()
		// Don't remove yet - caller needs it
	}()

	changelogPath := UploadParams.Changelog
	if UploadParams.Changelog == "" {
		var tmpChangelog *os.File
		tmpChangelog, err = os.CreateTemp(tmp, "wbtChangelog*.md")
		if err != nil {
			return
		}
		tempFiles = append(tempFiles, tmpChangelog.Name())
		defer func() {
			_ = tmpChangelog.Close()
			// Don't remove yet - caller needs it
		}()

		_, err = tmpChangelog.WriteString("No changelog provided")
		if err != nil {
			return
		}
		err = tmpChangelog.Sync()
		if err != nil {
			return
		}

		changelogPath = tmpChangelog.Name()
	}

	cLog := &changelog.Changelog{
		PreExistingFilePath: changelogPath,
		MarkupType:          changelog.MarkdownMT,
	}

	interfaceStringList := []string{}
	for _, i := range UploadParams.InterfaceVersions {
		interfaceStringList = append(interfaceStringList, fmt.Sprintf("%d", i))
	}

	interfaceString := strings.Join(interfaceStringList, ",")
	_, err = fmt.Fprintf(tmpToc, "## Interface: %s", interfaceString)
	if err != nil {
		return
	}
	err = tmpToc.Sync()
	if err != nil {
		return
	}

	tocFile, err := toc.NewToc(tmpToc.Name())
	if err != nil {
		return
	}

	prepPayload = &preparePayload{
		toc:       tocFile,
		changelog: cLog,
		cleanup: func() {
			for _, file := range tempFiles {
				_ = os.Remove(file)
			}
		},
	}
	return
}
