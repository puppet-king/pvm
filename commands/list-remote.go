package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"slices"

	"github.com/fatih/color"
)

var retrievePHPVersions = common.RetrievePHPVersions
var retrieveInstalledPHPVersions = common.RetrieveInstalledPHPVersions

func ListRemote() error {
	versions, err := retrievePHPVersions()
	if err != nil {
		return err
	}

	common.SortVersions(versions)

	installedVersions, _ := retrieveInstalledPHPVersions()

	currentVersion := common.GetCurrentVersionFolder()
	currentVersionNumber, currentVersionErr := common.ParseVersion(currentVersion, common.IsThreadSafeName(currentVersion), "")

	theme.Title("PHP versions available")
	for _, version := range versions {
		label := version.StringShort()

		idx := slices.IndexFunc(installedVersions, func(v common.Version) bool { return v.Same(version) })
		if idx != -1 {
			if currentVersionErr == nil && version.Same(currentVersionNumber) {
				label += " " + color.GreenString("[current]")
			} else {
				label += " " + color.CyanString("[installed]")
			}
		}

		color.White("    " + label)
	}

	return nil
}
