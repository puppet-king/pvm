package commands

import (
	"fmt"
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

type extensionStatusTag string

const (
	extensionTagEnabled   extensionStatusTag = "enabled"
	extensionTagDisabled  extensionStatusTag = "disabled"
	extensionTagAvailable extensionStatusTag = "available"
	extensionTagMissing   extensionStatusTag = "missing file"
)

type extensionListItem struct {
	Name string
	Tags []extensionStatusTag
}

type extensionInventory struct {
	Extensions     []extensionListItem
	ZendExtensions []extensionListItem
}

func Extensions(args []string) error {
	if len(args) < 1 {
		theme.Info("Usage: pvm extensions <list|ls|enable|disable> [extension[,extension...]]")
		return fmt.Errorf("you must specify an action")
	}

	currentVersion := common.GetCurrentVersionFolder()
	if currentVersion == "" {
		theme.Info("Select a PHP version with `pvm use <version>` first.")
		return fmt.Errorf("you do not have an active PHP version")
	}

	command := args[0]
	if command != "list" && command != "ls" && command != "enable" && command != "disable" {
		return fmt.Errorf("invalid action. must be 'list', 'ls', 'enable' or 'disable'")
	}

	paths, err := common.NewPVMPaths()
	if err != nil {
		return fmt.Errorf("could not determine your home directory: %w", err)
	}

	versionPath := paths.VersionDir(currentVersion)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		return fmt.Errorf("the specified version does not exist")
	}

	iniPath := filepath.Join(versionPath, "php.ini")
	ini, err := common.ReadPhpIni(iniPath)
	if err != nil {
		return fmt.Errorf("could not read php.ini for the active PHP version")
	}

	if command == "list" || command == "ls" {
		return listExtensions(versionPath, ini)
	}

	if len(args) < 2 {
		theme.Info("Usage: pvm extensions <enable|disable> <extension[,extension...]>")
		return fmt.Errorf("you must specify at least one extension")
	}

	extensions := normalizeExtensions(args[1])
	if len(extensions) == 0 {
		return fmt.Errorf("you must specify at least one extension")
	}

	newIni, results, changed := applyExtensionChanges(ini, command, extensions)
	if changed {
		if err := os.WriteFile(iniPath, []byte(newIni), 0644); err != nil {
			return fmt.Errorf("could not update php.ini for the active PHP version")
		}
	}

	reportExtensionResults(results)
	return nil
}

func listExtensions(versionPath string, ini string) error {
	extensionFiles, err := readAvailableExtensionFiles(versionPath)
	if err != nil {
		return fmt.Errorf("could not read the ext directory for the active PHP version")
	}

	inventory := buildExtensionInventory(ini, extensionFiles)
	reportExtensionGroup("Extensions", inventory.Extensions)
	reportExtensionGroup("Zend extensions", inventory.ZendExtensions)
	return nil
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
		matches := matchingDirectiveEntries(currentIni, extension)
		if len(matches) == 0 {
			results = append(results, extensionResult{
				name:    extension,
				kind:    "error",
				message: "Extension or Zend extension " + extension + " not found in php.ini",
			})
			continue
		}

		directiveKind := matches[0].Kind
		label := extensionLabel(extension, directiveKind)
		hasEnabled := false
		for _, match := range matches {
			if match.Enabled {
				hasEnabled = true
				break
			}
		}

		switch {
		case hasEnabled:
			if command == "enable" {
				results = append(results, extensionResult{
					name:    extension,
					kind:    "success",
					message: label + " is already enabled.",
				})
				continue
			}

			for _, match := range matches {
				splitIni[match.Line] = ";" + strings.TrimPrefix(splitIni[match.Line], ";")
			}
			currentIni = strings.Join(splitIni, separator)
			changed = true
			results = append(results, extensionResult{
				name:    extension,
				kind:    "success",
				message: label + " disabled.",
			})
		default:
			if command == "disable" {
				results = append(results, extensionResult{
					name:    extension,
					kind:    "success",
					message: label + " is already disabled.",
				})
				continue
			}

			for _, match := range matches {
				splitIni[match.Line] = strings.TrimPrefix(splitIni[match.Line], ";")
			}
			currentIni = strings.Join(splitIni, separator)
			changed = true
			results = append(results, extensionResult{
				name:    extension,
				kind:    "success",
				message: label + " enabled.",
			})
		}
	}

	if !changed {
		return currentIni, results, false
	}

	return currentIni, results, true
}

func matchingDirectiveEntries(ini string, extension string) []common.PhpIniExtension {
	normalizedExtension := common.NormalizeExtensionName(extension)
	matches := make([]common.PhpIniExtension, 0)

	for _, entry := range common.ParsePhpIniExtensions(ini) {
		if entry.Name != normalizedExtension {
			continue
		}

		matches = append(matches, entry)
	}

	return matches
}

func buildExtensionInventory(ini string, extensionFiles []string) extensionInventory {
	configuredExtensions := make(map[string]common.ExtensionStatus)
	configuredZendExtensions := make(map[string]common.ExtensionStatus)
	available := make(map[string]struct{}, len(extensionFiles))

	for _, entry := range common.ParsePhpIniExtensions(ini) {
		status := common.ExtensionDisabled
		if entry.Enabled {
			status = common.ExtensionEnabled
		}

		switch entry.Kind {
		case common.PhpIniZendExtensionDirective:
			mergeExtensionStatus(configuredZendExtensions, entry.Name, status)
		default:
			mergeExtensionStatus(configuredExtensions, entry.Name, status)
		}
	}

	for _, extensionFile := range extensionFiles {
		normalized := common.NormalizeExtensionName(extensionFile)
		if normalized == "" {
			continue
		}

		available[normalized] = struct{}{}
	}

	inventory := extensionInventory{
		Extensions:     make([]extensionListItem, 0),
		ZendExtensions: make([]extensionListItem, 0),
	}

	for name, status := range configuredExtensions {
		item := extensionListItem{Name: name, Tags: []extensionStatusTag{statusToTag(status)}}
		if _, ok := available[name]; !ok {
			item.Tags = append(item.Tags, extensionTagMissing)
		}
		inventory.Extensions = append(inventory.Extensions, item)
	}

	for name := range available {
		if _, ok := configuredExtensions[name]; ok {
			continue
		}
		if _, ok := configuredZendExtensions[name]; ok {
			continue
		}

		inventory.Extensions = append(inventory.Extensions, extensionListItem{
			Name: name,
			Tags: []extensionStatusTag{extensionTagAvailable},
		})
	}

	for name, status := range configuredZendExtensions {
		item := extensionListItem{Name: name, Tags: []extensionStatusTag{statusToTag(status)}}
		if _, ok := available[name]; !ok {
			item.Tags = append(item.Tags, extensionTagMissing)
		}
		inventory.ZendExtensions = append(inventory.ZendExtensions, item)
	}

	sortExtensionListItems(inventory.Extensions)
	sortExtensionListItems(inventory.ZendExtensions)

	return inventory
}

func mergeExtensionStatus(configured map[string]common.ExtensionStatus, name string, status common.ExtensionStatus) {
	if existingStatus, ok := configured[name]; ok && existingStatus == common.ExtensionEnabled {
		return
	}

	configured[name] = status
}

func statusToTag(status common.ExtensionStatus) extensionStatusTag {
	if status == common.ExtensionEnabled {
		return extensionTagEnabled
	}

	return extensionTagDisabled
}

func sortExtensionListItems(items []extensionListItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
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

func extensionLabel(name string, kind common.PhpIniDirectiveKind) string {
	if kind == common.PhpIniZendExtensionDirective {
		return "Zend extension " + name
	}

	return "Extension " + name
}

func reportExtensionGroup(title string, items []extensionListItem) {
	theme.Title(title)
	if len(items) == 0 {
		color.White("    none")
		return
	}

	for _, item := range items {
		color.White("    " + formatExtensionItem(item))
	}
}

func formatExtensionItem(item extensionListItem) string {
	parts := make([]string, 0, len(item.Tags)+1)
	parts = append(parts, item.Name)

	for _, tag := range item.Tags {
		parts = append(parts, formatExtensionTag(tag))
	}

	return strings.Join(parts, " ")
}

func formatExtensionTag(tag extensionStatusTag) string {
	label := fmt.Sprintf("[%s]", tag)

	switch tag {
	case extensionTagEnabled:
		return color.GreenString(label)
	case extensionTagDisabled:
		return color.YellowString(label)
	case extensionTagAvailable:
		return color.CyanString(label)
	case extensionTagMissing:
		return color.RedString(label)
	default:
		return label
	}
}
