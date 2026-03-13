package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type extensionResult struct {
	name    string
	kind    string
	message string
}

func Extensions(args []string) {
	if len(args) < 2 {
		theme.Error("You must specify an action and an extension.")
		theme.Info("Usage: pvm extensions <enable|disable> <extension>")
		return
	}

	currentVersion := common.GetCurrentVersionFolder()
	if currentVersion == "" {
		theme.Error("You do not have an active PHP version.")
		theme.Info("Select a PHP version with `pvm use <version>` first.")
		return
	}

	command := args[0]
	if command != "enable" && command != "disable" {
		theme.Error("Invalid action. Must be 'enable' or 'disable'.")
		return
	}

	extensions := normalizeExtensions(args[1])
	if len(extensions) == 0 {
		theme.Error("You must specify at least one extension.")
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

	iniPath := filepath.Join(versionPath, "php.ini")
	ini, err := common.ReadPhpIni(iniPath)
	if err != nil {
		theme.Error("Could not read php.ini for the active PHP version.")
		return
	}

	newIni, results, changed := applyExtensionChanges(ini, command, extensions)
	if changed {
		if err := os.WriteFile(iniPath, []byte(newIni), 0644); err != nil {
			theme.Error("Could not update php.ini for the active PHP version.")
			return
		}
	}

	reportExtensionResults(results)
}

func normalizeExtensions(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	extensions := make([]string, 0, len(parts))

	for _, part := range parts {
		extension := strings.TrimSpace(part)
		if extension == "" {
			continue
		}
		if _, ok := seen[extension]; ok {
			continue
		}

		seen[extension] = struct{}{}
		extensions = append(extensions, extension)
	}

	return extensions
}

func applyExtensionChanges(ini string, command string, extensions []string) (string, []extensionResult, bool) {
	separator := detectLineSeparator(ini)
	splitIni := regexp.MustCompile(`\r?\n`).Split(ini, -1)
	currentIni := ini
	results := make([]extensionResult, 0, len(extensions))
	changed := false

	for _, extension := range extensions {
		extensionStatus, lineNumber := common.GetExtensionStatus(currentIni, extension)

		switch extensionStatus {
		case common.ExtensionEnabled:
			if command == "enable" {
				results = append(results, extensionResult{
					name:    extension,
					kind:    "success",
					message: "Extension " + extension + " is already enabled.",
				})
				continue
			}

			splitIni[lineNumber] = ";" + splitIni[lineNumber]
			currentIni = strings.Join(splitIni, separator)
			changed = true
			results = append(results, extensionResult{
				name:    extension,
				kind:    "success",
				message: "Extension " + extension + " disabled.",
			})
		case common.ExtensionDisabled:
			if command == "disable" {
				results = append(results, extensionResult{
					name:    extension,
					kind:    "success",
					message: "Extension " + extension + " is already disabled.",
				})
				continue
			}

			splitIni[lineNumber] = strings.Replace(splitIni[lineNumber], ";", "", 1)
			currentIni = strings.Join(splitIni, separator)
			changed = true
			results = append(results, extensionResult{
				name:    extension,
				kind:    "success",
				message: "Extension " + extension + " enabled.",
			})
		default:
			results = append(results, extensionResult{
				name:    extension,
				kind:    "error",
				message: "Extension " + extension + " not found in php.ini",
			})
		}
	}

	if !changed {
		return currentIni, results, false
	}

	return currentIni, results, true
}

func detectLineSeparator(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}

	return "\n"
}

func reportExtensionResults(results []extensionResult) {
	for _, result := range results {
		if result.kind == "error" {
			theme.Error(result.message)
			continue
		}

		theme.Success(result.message)
	}
}
