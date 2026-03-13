package common

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Url        string
	ThreadSafe bool
}

func (v Version) Semantic() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

func (v Version) StringShort() string {
	semantic := v.Semantic()
	if v.ThreadSafe {
		return semantic
	}
	return semantic + " nts"
}

func (v Version) String() string {
	semantic := v.Semantic()
	if v.ThreadSafe {
		return semantic + " thread safe"
	}
	return semantic + " non-thread safe"
}

func ComputeVersion(text string, safe bool, url string) Version {
	versionRe := regexp.MustCompile(`([0-9]{1,3})(?:.([0-9]{1,3}))?(?:.([0-9]{1,3}))?`)
	matches := versionRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return Version{}
	}

	major, err := strconv.Atoi(matches[0][1])
	if err != nil {
		major = -1
	}

	minor, err := strconv.Atoi(matches[0][2])
	if err != nil {
		minor = -1
	}

	patch, err := strconv.Atoi(matches[0][3])
	if err != nil {
		patch = -1
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		ThreadSafe: safe,
		Url:        url,
	}
}

func (v Version) Compare(o Version) int {
	if v.Major == -1 || o.Major == -1 {
		return 0
	}
	if v.Major != o.Major {
		if v.Major < o.Major {
			return -1
		}
		return 1
	}

	if v.Minor == -1 || o.Minor == -1 {
		return 0
	}
	if v.Minor != o.Minor {
		if v.Minor < o.Minor {
			return -1
		}
		return 1
	}

	if v.Patch == -1 || o.Patch == -1 {
		return 0
	}
	if v.Patch != o.Patch {
		if v.Patch < o.Patch {
			return -1
		}
		return 1
	}

	return 0
}

func (v Version) CompareThreadSafe(o Version) int {
	result := v.Compare(o)
	if result != 0 {
		return result
	}

	if v.ThreadSafe == o.ThreadSafe {
		return 0
	}

	if v.ThreadSafe {
		return -1
	}
	return 1
}

func (v Version) LessThan(o Version) bool {
	return v.CompareThreadSafe(o) == -1
}

func (v Version) Same(o Version) bool {
	return v.CompareThreadSafe(o) == 0
}

func SortVersions(input []Version) []Version {
	sort.SliceStable(input, func(i, j int) bool {
		return input[i].LessThan(input[j])
	})
	return input
}

func RetrievePHPVersions() ([]Version, error) {
	resp, err := http.Get("https://downloads.php.net/~windows/releases/archives/")
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

		url := resolvedURL.String()
		name := match[2]

		if name != "" && len(name) > 15 && name[:15] == "php-devel-pack-" {
			continue
		}
		if name != "" && len(name) > 15 && name[:15] == "php-debug-pack-" {
			continue
		}
		if name != "" && len(name) > 15 && name[:14] == "php-test-pack-" {
			continue
		}
		if name != "" && strings.Contains(name, "src") {
			continue
		}
		if name != "" && !strings.HasSuffix(name, ".zip") {
			continue
		}

		threadSafe := true
		if name != "" && (strings.Contains(name, "nts") || strings.Contains(name, "NTS")) {
			threadSafe = false
		}
		if name != "" && !strings.Contains(name, "x64") {
			continue
		}

		versions = append(versions, ComputeVersion(name, threadSafe, url))
	}

	return versions, nil
}

func RetrieveInstalledPHPVersions() ([]Version, error) {
	versions := make([]Version, 0)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
		return versions, err
	}

	pvmPath := filepath.Join(homeDir, ".pvm")
	if _, err := os.Stat(pvmPath); os.IsNotExist(err) {
		return versions, errors.New("no PHP versions installed")
	}

	versionsPath := filepath.Join(pvmPath, "versions")
	if _, err := os.Stat(versionsPath); os.IsNotExist(err) {
		return versions, errors.New("no PHP versions installed")
	}

	folders, err := os.ReadDir(versionsPath)
	if err != nil {
		return versions, err
	}

	for _, folder := range folders {
		folderName := folder.Name()
		safe := true
		if strings.Contains(folderName, "nts") || strings.Contains(folderName, "NTS") {
			safe = false
		}

		versions = append(versions, ComputeVersion(folderName, safe, ""))
	}

	SortVersions(versions)
	return versions, nil
}

func GetCurrentVersionFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	content, err := os.ReadFile(filepath.Join(homeDir, ".pvm", "version"))
	if err != nil {
		return ""
	}

	return string(content)
}

func ReadPhpIni(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(file)
}

type ExtensionStatus int

const (
	ExtensionEnabled ExtensionStatus = iota + 1
	ExtensionDisabled
	ExtensionNotFound
)

func GetExtensionStatus(ini string, extension string) (ExtensionStatus, int) {
	lines := regexp.MustCompile(`\r?\n`).Split(ini, -1)

	for index, line := range lines {
		extensionMatches := regexp.MustCompile(`extension\s*=\s*["']?([^"']+)["']?`).FindStringSubmatch(line)
		if len(extensionMatches) == 0 {
			continue
		}

		if extensionMatches[1] == extension {
			noWhitespace := strings.TrimSpace(lines[index])
			if strings.HasPrefix(noWhitespace, ";") {
				return ExtensionDisabled, index
			}
			return ExtensionEnabled, index
		}
	}

	return ExtensionNotFound, -1
}
