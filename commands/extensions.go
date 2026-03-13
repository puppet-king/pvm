package commands

import (
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type extensionResult struct {
	name    string
	kind    string
	message string
}

type extensionInventory struct {
	Enabled   []string
	Disabled  []string
	Available []string
}

func Extensions(args []string) {
	if len(args) < 1 {
		theme.Error("You must specify an action.")
		theme.Info("Usage: pvm extensions <list|enable|disable> [extension[,extension...]]")
		return
	}

	currentVersion := common.GetCurrentVersionFolder()
	if currentVersion == "" {
		theme.Error("You do not have an active PHP version.")
		theme.Info("Select a PHP version with `pvm use <version>` first.")
		return
	}

	command := args[0]
	if command != "list" && command != "enable" && command != "disable" {
		theme.Error("Invalid action. Must be 'list', 'enable' or 'disable'.")
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

	if command == "list" {
		listExtensions(versionPath, ini)
		return
	}

	if len(args) < 2 {
		theme.Error("You must specify at least one extension.")
		theme.Info("Usage: pvm extensions <enable|disable> <extension[,extension...]>")
		return
	}

	extensions := normalizeExtensions(args[1])
	if len(extensions) == 0 {
		theme.Error("You must specify at least one extension.")
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

func listExtensions(versionPath string, ini string) {
	extensionFiles, err := readAvailableExtensionFiles(versionPath)
	if err != nil {
		theme.Error("Could not read the ext directory for the active PHP version.")
		return
	}

	inventory := buildExtensionInventory(ini, extensionFiles)

	reportExtensionGroup("Enabled extensions", inventory.Enabled)
	reportExtensionGroup("Disabled extensions", inventory.Disabled)
	reportExtensionGroup("Available extensions", inventory.Available)
}

func normalizeExtensions(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	extensions := make([]string, 0, len(parts))

	for _, part := range parts {
		extension := common.NormalizeExtensionName(part)
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

			splitIni[lineNumber] = ";" + strings.TrimPrefix(splitIni[lineNumber], ";")
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

			splitIni[lineNumber] = strings.TrimPrefix(splitIni[lineNumber], ";")
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

func buildExtensionInventory(ini string, extensionFiles []string) extensionInventory {
	configured := make(map[string]common.ExtensionStatus)
	available := make(map[string]struct{}, len(extensionFiles))

	for _, entry := range common.ParsePhpIniExtensions(ini) {
		status := common.ExtensionDisabled
		if entry.Enabled {
			status = common.ExtensionEnabled
		}

		if existingStatus, ok := configured[entry.Name]; ok && existingStatus == common.ExtensionEnabled {
			continue
		}

		configured[entry.Name] = status
	}

	for _, extensionFile := range extensionFiles {
		normalized := common.NormalizeExtensionName(extensionFile)
		if normalized == "" {
			continue
		}

		available[normalized] = struct{}{}
	}

	inventory := extensionInventory{
		Enabled:   make([]string, 0),
		Disabled:  make([]string, 0),
		Available: make([]string, 0),
	}

	for name, status := range configured {
		label := name
		if _, ok := available[name]; !ok {
			label += " (missing file)"
		}

		if status == common.ExtensionEnabled {
			inventory.Enabled = append(inventory.Enabled, label)
			continue
		}

		inventory.Disabled = append(inventory.Disabled, label)
	}

	for name := range available {
		if _, ok := configured[name]; ok {
			continue
		}

		inventory.Available = append(inventory.Available, name)
	}

	sort.Strings(inventory.Enabled)
	sort.Strings(inventory.Disabled)
	sort.Strings(inventory.Available)

	return inventory
}

func readAvailableExtensionFiles(versionPath string) ([]string, error) {
	extPath := filepath.Join(versionPath, "ext")
	entries, err := os.ReadDir(extPath)
	if err != nil {
		return nil, err
	}

	extensionFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".dll") {
			continue
		}

		extensionFiles = append(extensionFiles, entry.Name())
	}

	return extensionFiles, nil
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

func reportExtensionGroup(title string, extensions []string) {
	theme.Title(title)
	if len(extensions) == 0 {
		color.White("    none")
		return
	}

	for _, extension := range extensions {
		color.White("    " + extension)
	}
}
