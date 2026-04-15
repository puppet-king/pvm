package common

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func RetrievePHPVersions() ([]Version, error) {
	client := NewHTTPClient()
	resp, err := client.Get("https://downloads.php.net/~windows/releases/archives/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code while retrieving PHP versions: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParsePHPVersions(string(body), "https://downloads.php.net/~windows/releases/archives/")
}

func ParsePHPVersions(body string, baseURL string) ([]Version, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`(?i)<a\s+href="([^"]+)">([^<]+)</a>`)
	matches := re.FindAllStringSubmatch(body, -1)
	versions := make([]Version, 0)

	for _, match := range matches {
		resolvedURL, err := parsedBaseURL.Parse(match[1])
		if err != nil {
			continue
		}

		name := match[2]
		if !isSupportedArchiveName(name) {
			continue
		}

		version, err := ParseVersion(name, IsThreadSafeName(name), resolvedURL.String())
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func RetrieveInstalledPHPVersions() ([]Version, error) {
	paths, err := NewPVMPaths()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(paths.Root); os.IsNotExist(err) {
		return nil, errors.New("no PHP versions installed")
	}
	if _, err := os.Stat(paths.VersionsDir); os.IsNotExist(err) {
		return nil, errors.New("no PHP versions installed")
	}

	folders, err := os.ReadDir(paths.VersionsDir)
	if err != nil {
		return nil, err
	}

	versions := make([]Version, 0, len(folders))
	for _, folder := range folders {
		version, err := ParseVersion(folder.Name(), IsThreadSafeName(folder.Name()), "")
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	SortVersions(versions)
	return versions, nil
}

func GetCurrentVersionFolder() string {
	paths, err := NewPVMPaths()
	if err != nil {
		return ""
	}

	content, err := os.ReadFile(paths.CurrentVersionFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func ReadPhpIni(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func isSupportedArchiveName(name string) bool {
	if !strings.HasSuffix(name, ".zip") {
		return false
	}
	if strings.Contains(name, "src") {
		return false
	}
	if strings.HasPrefix(name, "php-devel-pack-") {
		return false
	}
	if strings.HasPrefix(name, "php-debug-pack-") {
		return false
	}
	if strings.HasPrefix(name, "php-test-pack-") {
		return false
	}

	return strings.Contains(name, "x64")
}
