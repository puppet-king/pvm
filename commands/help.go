package commands

import (
	"fmt"
	"hjbdev/pvm/theme"

	"github.com/fatih/color"
)

var version = "dev"

func Help(notFoundError bool) {
	theme.Title("pvm: PHP Version Manager")
	theme.Info(fmt.Sprintf("Version %s", version))

	if notFoundError {
		theme.Error("Command not found")
	}

	fmt.Println("Available Commands:")
	printHelpCommand("extensions <list|ls|enable|disable> [extension[,extension...]]", "e")
	printHelpCommand("help")
	printHelpCommand("install", "i")
	printHelpCommand("list [remote]", "ls")
	printHelpCommand("bin")
	printHelpCommand("use <version>", "u")
}

func printHelpCommand(command string, aliases ...string) {
	line := "    " + command
	if len(aliases) == 0 {
		fmt.Println(line)
		return
	}

	fmt.Println(line + " " + color.HiBlackString("(alias: %s)", aliases[0]))
}
