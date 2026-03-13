package commands

import "hjbdev/pvm/common"

func findExactVersion(versions []common.Version, spec common.VersionSpec) (common.Version, bool) {
	for _, version := range versions {
		if spec.Matches(version) && spec.HasMinor && spec.HasPatch {
			return version, true
		}
	}

	return common.Version{}, false
}

func findLatestPatch(versions []common.Version, spec common.VersionSpec) (common.Version, bool) {
	var selected common.Version
	found := false

	for _, version := range versions {
		if !spec.Matches(version) || !spec.HasMinor {
			continue
		}
		if !found || selected.LessThan(version) {
			selected = version
			found = true
		}
	}

	return selected, found
}

func findLatestMinor(versions []common.Version, spec common.VersionSpec) (common.Version, bool) {
	var selected common.Version
	found := false

	for _, version := range versions {
		if !spec.Matches(version) {
			continue
		}
		if !found || selected.LessThan(version) {
			selected = version
			found = true
		}
	}

	return selected, found
}

func resolveInstallVersion(versions []common.Version, spec common.VersionSpec) (common.Version, bool) {
	if spec.HasMinor && spec.HasPatch {
		return findExactVersion(versions, spec)
	}
	if spec.HasMinor {
		return findLatestPatch(versions, spec)
	}

	return findLatestMinor(versions, spec)
}
