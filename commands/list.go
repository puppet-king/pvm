package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"strings"

	"github.com/fatih/color"
)

func List() {
	versions, err := common.RetrieveInstalledPHPVersions()
	if err != nil {
		theme.Error(err.Error())
	}

	theme.Title("Installed PHP versions")

	currentVersion := common.GetCurrentVersionFolder()
	currentVersionSafe := !strings.Contains(strings.ToLower(currentVersion), "nts")
	currentVersionNumber := common.ComputeVersion(currentVersion, currentVersionSafe, "")

	for _, version := range versions {
		label := version.StringShort()
		if currentVersion != "" && version.Same(currentVersionNumber) {
			label += " (current)"
		}

		color.White("    " + label)
	}
}
