package commands

import (
	"fmt"
	"hjbdev/pvm/theme"
)

func List(args []string) error {
	action, ok := parseListAction(args)
	if !ok {
		theme.Error("Invalid list action.")
		theme.Info("Usage: pvm list [remote]")
		return fmt.Errorf("invalid list action")
	}

	switch action {
	case "remote":
		return ListRemote()
	default:
		return ListLocal()
	}
}

func parseListAction(args []string) (string, bool) {
	if len(args) == 0 {
		return "local", true
	}

	if len(args) == 1 && args[0] == "remote" {
		return "remote", true
	}

	return "", false
}
