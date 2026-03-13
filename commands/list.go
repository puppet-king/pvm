package commands

import "hjbdev/pvm/theme"

func List(args []string) {
	action, ok := parseListAction(args)
	if !ok {
		theme.Error("Invalid list action.")
		theme.Info("Usage: pvm list [remote]")
		return
	}

	switch action {
	case "remote":
		ListRemote()
	default:
		ListLocal()
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
