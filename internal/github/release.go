package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/McTalian/wow-build-tools/internal/logger"
)

type GetReleaseArgs struct {
	Slug string
	Tag  string
}

var GetReleaseParams = &GetReleaseArgs{}

type GitHubRelease struct {
	GitHubReleasePayload
	Id   int `json:"id"`
	Slug string
}

type GitHubReleasePayload struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func (r *GitHubRelease) UploadAsset(fileName string, filePath string, logGroup *logger.LogGroup) error {
	return UploadGitHubAsset(r.Slug, r.Id, fileName, filePath, logGroup)
}

func (r *GitHubRelease) getPayload() (*bytes.Buffer, error) {
	payload, err := json.Marshal(&r.GitHubReleasePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal release: %w", err)
	}

	return bytes.NewBuffer(payload), nil
}

func (r *GitHubRelease) UpdateRelease(newPayload GitHubReleasePayload) error {
	url := fmt.Sprintf("%srepos/%s/releases/%d", githubApiUrl, r.Slug, r.Id)

	r.GitHubReleasePayload = newPayload
	body, err := r.getPayload()
	if err != nil {
		return fmt.Errorf("failed to marshal release: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := clientRequest(req)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update release: %d", resp.StatusCode)
	}

	return nil
}

func CreateRelease(slug string, payload GitHubReleasePayload) (release *GitHubRelease, err error) {
	r := &GitHubRelease{
		GitHubReleasePayload: payload,
		Slug:                 slug,
	}

	url := fmt.Sprintf("%srepos/%s/releases", githubApiUrl, r.Slug)

	body, err := r.getPayload()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal release: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := clientRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create release: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(release)

	return release, err
}

var ErrReleaseNotFound = fmt.Errorf("release not found")

func GetRelease(slug, tag string) (release *GitHubRelease, err error) {
	url := fmt.Sprintf("%srepos/%s/releases/tags/%s", githubApiUrl, slug, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	resp, err := clientRequest(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		err = ErrReleaseNotFound
		return
	}

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&release)
		if err != nil {
			return
		}
		release.Slug = slug

		return
	}

	return
}

func RunReleaseGet(slug, tag string) error {
	release, err := GetRelease(slug, tag)
	if err != nil {
		logger.Error("Failed to get release ID")
		return err
	}

	logger.Info("Release ID: %d", release.Id)
	logger.Info("Tag Name: %s", release.TagName)
	logger.Info("Name: %s", release.Name)
	logger.Info("Draft: %t", release.Draft)
	logger.Info("Prerelease: %t", release.Prerelease)
	logger.Info("%s", release.Body)
	return nil
}
