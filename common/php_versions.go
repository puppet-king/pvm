package common

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
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

	// Parse the HTML table to get file info (name, size)
	// Format: <tr><td>icon</td><td><a href="file.zip">file.zip</a></td><td>date</td><td>size</td></tr>
	fileSizes := make(map[string]int64)

	// Match table rows: look for <tr>...</tr> and extract href, name, and size
	// The size is in the 4th <td> element
	rowRe := regexp.MustCompile(`(?si)<tr[^>]*>.*?<td[^>]*>.*?</td>\s*<td[^>]*>.*?<a\s+href="([^"]+)">([^<]+)</a>.*?</td>\s*<td[^>]*>([^<]*)</td>\s*<td[^>]*>([^<]*)</td>.*?</tr>`)
	matches := rowRe.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		if len(match) >= 5 {
			name := match[2]
			sizeStr := strings.TrimSpace(match[4])
			if sizeBytes := parseSize(sizeStr); sizeBytes > 0 {
				fileSizes[name] = sizeBytes
			}
		}
	}

	// Also use the original link matching for fallback
	linkRe := regexp.MustCompile(`(?i)<a\s+href="([^"]+)">([^<]+)</a>`)
	linkMatches := linkRe.FindAllStringSubmatch(body, -1)

	versions := make([]Version, 0)

	for _, match := range linkMatches {
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

		// Add size if available
		if size, ok := fileSizes[name]; ok {
			version.SizeBytes = size
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// parseSize parses human-readable size strings like "16M", "1.5G", "512K" to bytes
func parseSize(sizeStr string) int64 {
	if sizeStr == "" || sizeStr == "-" {
		return 0
	}

	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Try to parse number with unit
	re := regexp.MustCompile(`^([\d.]+)\s*([KMGT]?)B?$`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) < 3 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := matches[2]
	multiplier := int64(1)

	switch unit {
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier))
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
