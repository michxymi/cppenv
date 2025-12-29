package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type pypiResponse struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

// GetLatestVersion queries PyPI for the latest version of a package
func GetLatestVersion(packageName string) (string, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to query PyPI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("package not found on PyPI: %s (status %d)", packageName, resp.StatusCode)
	}

	var data pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to parse PyPI response: %w", err)
	}

	if data.Info.Version == "" {
		return "", fmt.Errorf("no version found for package: %s", packageName)
	}

	return data.Info.Version, nil
}
