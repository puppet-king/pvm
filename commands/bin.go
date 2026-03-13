package commands

import (
	"fmt"
	"hjbdev/pvm/theme"

	"hjbdev/pvm/common"
)

func Bin() error {
	theme.Title("pvm: PHP Version Manager")

	paths, err := common.NewPVMPaths()
	if err != nil {
		return err
	}

	fmt.Println("Add the following directory to your PATH:")
	fmt.Println("    " + paths.BinDir)
	return nil
}
