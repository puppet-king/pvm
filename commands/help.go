package commands

import (
	"fmt"
	"hjbdev/pvm/theme"
)

var version = "dev"

func Help(notFoundError bool) {
	theme.Title("pvm: PHP Version Manager")
	theme.Info(fmt.Sprintf("Version %s", version))

	if notFoundError {
		theme.Error("Command not found")
	}

	fmt.Println("Available Commands:")
	fmt.Println("    help")
	fmt.Println("    install")
	fmt.Println("    list")
	fmt.Println("    list-remote")
	fmt.Println("    path")
	fmt.Println("    use")
}
