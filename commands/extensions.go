package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Extensions(args []string) {
	if len(args) < 2 {
		theme.Error("You must specify an action and an extension.")
		theme.Info("Usage: pvm extensions <enable|disable> <extension>")
		return
	}

	// determine which version is currently selected
	currentVersion := common.GetCurrentVersionFolder()

	if currentVersion == "" {
		theme.Error("You do not have an active PHP version.")
		theme.Info("Select a PHP version with `pvm use <version>` first.")
		return
	}

	command := args[0]
	ext := args[1]

	if command != "enable" && command != "disable" {
		theme.Error("Invalid action. Must be 'enable' or 'disable'.")
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		theme.Error("Could not determine your home directory.")
		return
	}

	versionPath := filepath.Join(homeDir, ".pvm", "versions", currentVersion)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		theme.Error("The specified version does not exist.")
		return
	}

	extensions := strings.Split(ext, ",")

	for _, extension := range extensions {
		handleExtension(extension, command, homeDir, currentVersion)
	}
}

func handleExtension(ext string, command string, homeDir string, currentVersion string) {
	iniPath := filepath.Join(homeDir, ".pvm", "versions", currentVersion, "php.ini")
	ini, err := common.ReadPhpIni(iniPath)
	if err != nil {
		theme.Error("Could not read php.ini for the active PHP version.")
		return
	}

	splitIni := regexp.MustCompile(`\r?\n`).Split(ini, -1)
	extensionStatus, lineNumber := common.GetExtensionStatus(ini, ext)

	switch extensionStatus {
	case common.ExtensionEnabled:
		if command == "enable" {
			theme.Success("Extension " + ext + " is already enabled.")
		} else {
			disabledLine := ";" + splitIni[lineNumber]
			splitIni[lineNumber] = disabledLine
			newIni := strings.Join(splitIni, "\n")
			if err := os.WriteFile(iniPath, []byte(newIni), 0644); err != nil {
				theme.Error("Could not update php.ini for the active PHP version.")
				return
			}
			theme.Success("Extension " + ext + " disabled.")
		}
	case common.ExtensionDisabled:
		if command == "enable" {
			enabledLine := strings.Replace(splitIni[lineNumber], ";", "", 1)
			splitIni[lineNumber] = enabledLine
			newIni := strings.Join(splitIni, "\n")
			if err := os.WriteFile(iniPath, []byte(newIni), 0644); err != nil {
				theme.Error("Could not update php.ini for the active PHP version.")
				return
			}
			theme.Success("Extension " + ext + " enabled.")
		} else {
			theme.Success("Extension " + ext + " is already disabled.")
		}
	default:
		theme.Error("Extension " + ext + " not found in php.ini")
	}
}
