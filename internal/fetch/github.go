package fetch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func DownloadBlobFromGitHub(ctx context.Context, url string) ([]byte, error) {
	url = strings.Replace(url, "https://gitHub.com", "https://raw.githubusercontent.com", 1)
	url = strings.Replace(url, "/blob/", "/", 1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for URL '%s': %w", url, err)
	}

	if tokenFromEnv := os.Getenv("GITHUB_TOKEN"); tokenFromEnv != "" {
		req.Header.Set("Authorization", "Bearer: "+tokenFromEnv)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching blob '%s' from GitHub: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching blob '%s' from GitHub: %w", url, err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body from GitHub response '%s': %w", url, err)
	}

	return body, nil
}

type GitHubAuth struct {
	Token string
}

type GitHubLicense struct {
	Filename string `json:"name"`
	URL      string `json:"html_url"`
	Content  []byte `json:"content"`
	Sha      string `json:"sha"`
}

func DownloadLicenseFromGithub(ctx context.Context, repoURL string, ref string) (*GitHubLicense, error) {
	url := strings.Replace(repoURL, "https://github.com", "https://api.github.com/repos", 1) + "/license?ref=" + ref

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for URL '%s': %w", url, err)
	}

	if tokenFromEnv := os.Getenv("GITHUB_TOKEN"); tokenFromEnv != "" {
		req.Header.Set("Authorization", "Bearer: "+tokenFromEnv)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching license from GitHub at '%s': %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return DownloadLicenseFromGithub(ctx, repoURL, "master")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching license from GitHub at '%s': %v %v", url, res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body from GitHub response '%s': %w", url, err)
	}

	var ghlicense GitHubLicense
	err = json.Unmarshal(body, &ghlicense)
	if err != nil {
		return nil, fmt.Errorf("error parsing GitHub response JSON (url: '%s'): %w", url, err)
	}

	var licenseText []byte
	_, err = base64.StdEncoding.Decode(licenseText, ghlicense.Content)
	if err != nil {
		return nil, fmt.Errorf("error parsing GitHub license content base64 (url: '%s'): %w", url, err)
	}

	ghlicense.Content = licenseText

	return &ghlicense, nil
}
