package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const githubAPIBaseURL = "https://api.github.com"

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestRelease(repoOwner, repoName string) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBaseURL, repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func SelectVersionByRuntime(release *GitHubRelease) (string, error) {
	currentOS := runtime.GOOS
	arch := "amd64"
	var platform string

	// TODO: Add support for Windows
	//case "windows":
	//	platform = "windows"
	switch currentOS {
	case "darwin":
		platform = "darwin"
	case "linux":
		platform = "linux"
	}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, platform) && strings.Contains(asset.Name, arch) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no assets found for %s %s", platform, arch)
}

func DownloadRelease(url, fileName string) (string, error) {
	out, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download release: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	fmt.Printf("Release downloaded as %s\n", fileName)
	tarPath := fmt.Sprintf("%s/%s", currentDir, fileName)

	return tarPath, nil
}
