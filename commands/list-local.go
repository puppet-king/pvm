package commands

import (
	"fmt"
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"

	"github.com/fatih/color"
)

func ListLocal() error {
	versions, err := common.RetrieveInstalledPHPVersions()
	if err != nil {
		return err
	}

	theme.Title("Installed PHP versions")

	currentVersion := common.GetCurrentVersionFolder()
	currentVersionNumber, currentVersionErr := common.ParseVersion(currentVersion, common.IsThreadSafeName(currentVersion), "")

	for _, version := range versions {
		label := version.StringShort()
		if currentVersionErr == nil && version.Same(currentVersionNumber) {
			label += " " + color.GreenString("[current]")
		}

		color.White("    " + label)
	}

	if currentVersion != "" && currentVersionErr != nil {
		return fmt.Errorf("invalid current version metadata: %w", currentVersionErr)
	}

	return nil
}
