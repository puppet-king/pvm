package common

import (
	"fmt"
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
	SizeBytes  int64
}

type VersionSpec struct {
	Major      int
	Minor      int
	Patch      int
	HasMinor   bool
	HasPatch   bool
	ThreadSafe bool
}

var (
	versionRe     = regexp.MustCompile(`\b([0-9]{1,3})(?:\.([0-9]{1,3}))?(?:\.([0-9]{1,3}))?\b`)
	fullVersionRe = regexp.MustCompile(`\b([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\b`)
)

func (v Version) Semantic() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
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

func (v Version) Compare(o Version) int {
	switch {
	case v.Major < o.Major:
		return -1
	case v.Major > o.Major:
		return 1
	case v.Minor < o.Minor:
		return -1
	case v.Minor > o.Minor:
		return 1
	case v.Patch < o.Patch:
		return -1
	case v.Patch > o.Patch:
		return 1
	default:
		return 0
	}
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

func (s VersionSpec) Matches(version Version) bool {
	if s.ThreadSafe != version.ThreadSafe {
		return false
	}
	if s.Major != version.Major {
		return false
	}
	if s.HasMinor && s.Minor != version.Minor {
		return false
	}
	if s.HasPatch && s.Patch != version.Patch {
		return false
	}

	return true
}

func ParseVersion(text string, threadSafe bool, url string) (Version, error) {
	matches := fullVersionRe.FindStringSubmatch(text)
	if len(matches) == 0 {
		return Version{}, fmt.Errorf("invalid version: %s", text)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, err
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, err
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Version{}, err
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		ThreadSafe: threadSafe,
		Url:        url,
	}, nil
}

func ParseVersionSpec(text string, threadSafe bool) (VersionSpec, error) {
	matches := versionRe.FindStringSubmatch(text)
	if len(matches) == 0 {
		return VersionSpec{}, fmt.Errorf("invalid version: %s", text)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return VersionSpec{}, err
	}

	spec := VersionSpec{
		Major:      major,
		ThreadSafe: threadSafe,
	}

	if matches[2] != "" {
		minor, err := strconv.Atoi(matches[2])
		if err != nil {
			return VersionSpec{}, err
		}
		spec.Minor = minor
		spec.HasMinor = true
	}

	if matches[3] != "" {
		patch, err := strconv.Atoi(matches[3])
		if err != nil {
			return VersionSpec{}, err
		}
		spec.Patch = patch
		spec.HasPatch = true
	}

	return spec, nil
}

func SortVersions(input []Version) []Version {
	sort.SliceStable(input, func(i, j int) bool {
		return input[i].LessThan(input[j])
	})

	return input
}

func IsThreadSafeName(name string) bool {
	return !strings.Contains(strings.ToLower(name), "nts")
}
