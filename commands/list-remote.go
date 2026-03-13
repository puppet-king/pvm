package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"slices"

	"github.com/fatih/color"
)

func ListRemote() error {
	versions, err := common.RetrievePHPVersions()
	if err != nil {
		return err
	}

	common.SortVersions(versions)

	installedVersions, _ := common.RetrieveInstalledPHPVersions()

	theme.Title("PHP versions available")
	for _, version := range versions {
		idx := slices.IndexFunc(installedVersions, func(v common.Version) bool { return v.Same(version) })
		found := " "
		if idx != -1 {
			found = "*"
		}
		color.White(found + "   " + version.StringShort())
	}

	return nil
}
