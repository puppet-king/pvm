package common

import (
	"path"
	"regexp"
	"strings"
)

type ExtensionStatus int

type PhpIniDirectiveKind string

type PhpIniExtension struct {
	Name    string
	Enabled bool
	Line    int
	Kind    PhpIniDirectiveKind
}

const (
	PhpIniExtensionDirective     PhpIniDirectiveKind = "extension"
	PhpIniZendExtensionDirective PhpIniDirectiveKind = "zend_extension"

	ExtensionEnabled ExtensionStatus = iota + 1
	ExtensionDisabled
	ExtensionNotFound
)

func GetExtensionStatus(ini string, extension string) (ExtensionStatus, int) {
	normalizedExtension := NormalizeExtensionName(extension)

	for _, parsedExtension := range ParsePhpIniExtensions(ini) {
		if parsedExtension.Kind != PhpIniExtensionDirective {
			continue
		}
		if parsedExtension.Name != normalizedExtension {
			continue
		}

		if parsedExtension.Enabled {
			return ExtensionEnabled, parsedExtension.Line
		}

		return ExtensionDisabled, parsedExtension.Line
	}

	return ExtensionNotFound, -1
}

func GetDirectiveStatus(ini string, extension string) (ExtensionStatus, int, PhpIniDirectiveKind) {
	normalizedExtension := NormalizeExtensionName(extension)

	for _, parsedExtension := range ParsePhpIniExtensions(ini) {
		if parsedExtension.Name != normalizedExtension {
			continue
		}

		if parsedExtension.Enabled {
			return ExtensionEnabled, parsedExtension.Line, parsedExtension.Kind
		}

		return ExtensionDisabled, parsedExtension.Line, parsedExtension.Kind
	}

	return ExtensionNotFound, -1, ""
}

func ParsePhpIniExtensions(ini string) []PhpIniExtension {
	lines := regexp.MustCompile(`\r?\n`).Split(ini, -1)
	directiveRe := regexp.MustCompile(`^\s*(;?)\s*(zend_extension|extension)\s*=\s*(?:"([^"]+)"|'([^']+)'|([^;\s]+))\s*(?:;.*)?$`)
	extensions := make([]PhpIniExtension, 0)

	for index, line := range lines {
		matches := directiveRe.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		rawValue := matches[3]
		if rawValue == "" {
			rawValue = matches[4]
		}
		if rawValue == "" {
			rawValue = matches[5]
		}

		name := NormalizeExtensionName(rawValue)
		if name == "" {
			continue
		}

		extensions = append(extensions, PhpIniExtension{
			Name:    name,
			Enabled: matches[1] != ";",
			Line:    index,
			Kind:    PhpIniDirectiveKind(matches[2]),
		})
	}

	return extensions
}

func NormalizeExtensionName(value string) string {
	normalized := strings.TrimSpace(value)

	if idx := strings.Index(normalized, ";"); idx != -1 {
		normalized = strings.TrimSpace(normalized[:idx])
	}

	normalized = strings.Trim(normalized, `"'`)
	if normalized == "" {
		return ""
	}

	normalized = strings.ReplaceAll(normalized, `\`, "/")
	normalized = strings.ToLower(path.Base(normalized))
	normalized = strings.TrimSuffix(normalized, ".dll")
	normalized = strings.TrimPrefix(normalized, "php_")

	return strings.TrimSpace(normalized)
}
