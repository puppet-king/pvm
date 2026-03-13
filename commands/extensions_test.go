package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeExtensions_TrimsSkipsEmptyAndDeduplicates(t *testing.T) {
	extensions := normalizeExtensions("curl, ext\\php_mbstring.dll ,, curl,openssl ")

	assert.Equal(t, []string{"curl", "mbstring", "openssl"}, extensions)
}

func TestApplyExtensionChanges_PreservesCRLFAndAppliesMultipleChanges(t *testing.T) {
	ini := "extension=curl\r\n;extension=openssl\r\nextension=mysqli\r\n"

	updated, results, changed := applyExtensionChanges(ini, "disable", []string{"curl", "openssl", "missing"})

	assert.True(t, changed)
	assert.Equal(t, ";extension=curl\r\n;extension=openssl\r\nextension=mysqli\r\n", updated)
	assert.Equal(t, []extensionResult{
		{name: "curl", kind: "success", message: "Extension curl disabled."},
		{name: "openssl", kind: "success", message: "Extension openssl is already disabled."},
		{name: "missing", kind: "error", message: "Extension or Zend extension missing not found in php.ini"},
	}, results)
}

func TestApplyExtensionChanges_TogglesZendExtensions(t *testing.T) {
	ini := joinLines(
		`;zend_extension=php_xdebug.dll`,
		`zend_extension=php_opcache.dll`,
	)

	updated, results, changed := applyExtensionChanges(ini, "enable", []string{"xdebug", "opcache"})

	assert.True(t, changed)
	assert.Equal(t, joinLines(
		`zend_extension=php_xdebug.dll`,
		`zend_extension=php_opcache.dll`,
	), updated)
	assert.Equal(t, []extensionResult{
		{name: "xdebug", kind: "success", message: "Zend extension xdebug enabled."},
		{name: "opcache", kind: "success", message: "Zend extension opcache is already enabled."},
	}, results)
}

func TestApplyExtensionChanges_DisablesAllDuplicateDirectives(t *testing.T) {
	ini := joinLines(
		`extension=php_curl.dll`,
		`;extension=php_curl.dll`,
		`extension=php_curl.dll`,
	)

	updated, results, changed := applyExtensionChanges(ini, "disable", []string{"curl"})

	assert.True(t, changed)
	assert.Equal(t, joinLines(
		`;extension=php_curl.dll`,
		`;extension=php_curl.dll`,
		`;extension=php_curl.dll`,
	), updated)
	assert.Equal(t, []extensionResult{{name: "curl", kind: "success", message: "Extension curl disabled."}}, results)
}

func TestApplyExtensionChanges_EnablesAllDuplicateZendDirectives(t *testing.T) {
	ini := joinLines(
		`;zend_extension=php_xdebug.dll`,
		`;zend_extension=php_xdebug.dll`,
	)

	updated, results, changed := applyExtensionChanges(ini, "enable", []string{"xdebug"})

	assert.True(t, changed)
	assert.Equal(t, joinLines(
		`zend_extension=php_xdebug.dll`,
		`zend_extension=php_xdebug.dll`,
	), updated)
	assert.Equal(t, []extensionResult{{name: "xdebug", kind: "success", message: "Zend extension xdebug enabled."}}, results)
}

func TestExtensions_UpdatesPhpIniOnceForMultipleExtensions(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	phpIniPath := writeActivePhpVersion(t, homeDir, "8.3.1", "extension=curl\r\n;extension=openssl\r\n")

	Extensions([]string{"disable", "curl, openssl"})

	contents, err := os.ReadFile(phpIniPath)
	require.NoError(t, err)
	assert.Equal(t, ";extension=curl\r\n;extension=openssl\r\n", string(contents))
}

func TestExtensions_HandlesMissingPhpIniGracefully(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	writeActiveVersionMetadata(t, homeDir, "8.3.1")
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pvm", "versions", "8.3.1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pvm", "versions", "8.3.1", "ext"), 0755))

	assert.NotPanics(t, func() {
		Extensions([]string{"enable", "curl"})
	})

	_, err := os.Stat(filepath.Join(homeDir, ".pvm", "versions", "8.3.1", "php.ini"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestExtensions_AppliesKnownExtensionsEvenWhenSomeAreMissing(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	phpIniPath := writeActivePhpVersion(t, homeDir, "8.3.1", ";extension=curl\n;extension=openssl\n")

	Extensions([]string{"enable", "curl,missing,openssl"})

	contents, err := os.ReadFile(phpIniPath)
	require.NoError(t, err)
	assert.Equal(t, "extension=curl\nextension=openssl\n", string(contents))
}

func TestExtensions_TogglesZendExtensionFromCli(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	phpIniPath := writeActivePhpVersion(t, homeDir, "8.3.1", joinLines(
		`;zend_extension=php_xdebug.dll`,
		`extension=php_curl.dll`,
	))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pvm", "versions", "8.3.1", "ext", "php_xdebug.dll"), []byte(""), 0644))

	Extensions([]string{"enable", "xdebug"})

	contents, err := os.ReadFile(phpIniPath)
	require.NoError(t, err)
	assert.Equal(t, joinLines(
		`zend_extension=php_xdebug.dll`,
		`extension=php_curl.dll`,
	), string(contents))
}

func TestExtensions_ListAliasMatchesListAction(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	writeActivePhpVersion(t, homeDir, "8.3.1", joinLines(
		`extension=php_curl.dll`,
		`;zend_extension=php_xdebug.dll`,
	))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pvm", "versions", "8.3.1", "ext", "php_xdebug.dll"), []byte(""), 0644))

	listOutput := captureStdout(t, func() {
		Extensions([]string{"list"})
	})
	lsOutput := captureStdout(t, func() {
		Extensions([]string{"ls"})
	})

	assert.Equal(t, listOutput, lsOutput)
	assert.Contains(t, lsOutput, "Extensions")
	assert.Contains(t, lsOutput, "Zend extensions")
	assert.Contains(t, lsOutput, "curl")
	assert.Contains(t, lsOutput, "xdebug")
}

func TestExtensions_DoesNotTouchPhpIniWithoutActiveVersionMetadata(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	versionDir := filepath.Join(homeDir, ".pvm", "versions", "8.3.1")
	require.NoError(t, os.MkdirAll(filepath.Join(versionDir, "ext"), 0755))
	phpIniPath := filepath.Join(versionDir, "php.ini")
	require.NoError(t, os.WriteFile(phpIniPath, []byte(";extension=curl\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "ext", "php_curl.dll"), []byte(""), 0644))

	Extensions([]string{"enable", "curl"})

	contents, err := os.ReadFile(phpIniPath)
	require.NoError(t, err)
	assert.Equal(t, ";extension=curl\n", string(contents))
}

func TestBuildExtensionInventory_SeparatesRegularAndZendExtensions(t *testing.T) {
	ini := joinLines(
		`extension=php_curl.dll`,
		`;extension="ext\\php_openssl.dll"`,
		`extension=missing`,
		`zend_extension=php_xdebug.dll`,
		`; zend_extension = php_opcache.dll`,
	)

	inventory := buildExtensionInventory(ini, []string{"php_curl.dll", "php_mbstring.dll", "php_openssl.dll", "php_opcache.dll"})

	assert.Equal(t, extensionInventory{
		Extensions: []extensionListItem{
			{Name: "curl", Tags: []extensionStatusTag{extensionTagEnabled}},
			{Name: "mbstring", Tags: []extensionStatusTag{extensionTagAvailable}},
			{Name: "missing", Tags: []extensionStatusTag{extensionTagEnabled, extensionTagMissing}},
			{Name: "openssl", Tags: []extensionStatusTag{extensionTagDisabled}},
		},
		ZendExtensions: []extensionListItem{
			{Name: "opcache", Tags: []extensionStatusTag{extensionTagDisabled}},
			{Name: "xdebug", Tags: []extensionStatusTag{extensionTagEnabled, extensionTagMissing}},
		},
	}, inventory)
}

func TestBuildExtensionInventory_PrefersEnabledStateWhenExtensionAppearsTwice(t *testing.T) {
	ini := joinLines(
		`;extension=php_curl.dll`,
		`extension=curl`,
		`;zend_extension=php_xdebug.dll`,
		`zend_extension=php_xdebug.dll`,
	)

	inventory := buildExtensionInventory(ini, []string{"php_curl.dll", "php_xdebug.dll"})

	assert.Equal(t, []extensionListItem{{Name: "curl", Tags: []extensionStatusTag{extensionTagEnabled}}}, inventory.Extensions)
	assert.Equal(t, []extensionListItem{{Name: "xdebug", Tags: []extensionStatusTag{extensionTagEnabled}}}, inventory.ZendExtensions)
}

func TestReadAvailableExtensionFiles_FiltersNonDllEntries(t *testing.T) {
	versionPath := t.TempDir()
	extDir := filepath.Join(versionPath, "ext")
	require.NoError(t, os.MkdirAll(extDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(extDir, "php_curl.dll"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(extDir, "readme.txt"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "nested"), 0755))

	files, err := readAvailableExtensionFiles(versionPath)

	require.NoError(t, err)
	assert.Equal(t, []string{"php_curl.dll"}, files)
}

func TestFormatExtensionItem_FormatsNameBeforeTags(t *testing.T) {
	previousNoColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() {
		color.NoColor = previousNoColor
	})

	formatted := formatExtensionItem(extensionListItem{
		Name: "xdebug",
		Tags: []extensionStatusTag{extensionTagEnabled, extensionTagMissing},
	})

	assert.Equal(t, "xdebug [enabled] [missing file]", formatted)
}

func setHomeDir(t *testing.T, homeDir string) {
	t.Helper()
	t.Setenv("HOME", homeDir)
}

func writeActiveVersionMetadata(t *testing.T, homeDir string, version string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pvm"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pvm", "version"), []byte(version), 0644))
}

func writeActivePhpVersion(t *testing.T, homeDir string, version string, ini string) string {
	t.Helper()
	writeActiveVersionMetadata(t, homeDir, version)
	versionDir := filepath.Join(homeDir, ".pvm", "versions", version)
	require.NoError(t, os.MkdirAll(filepath.Join(versionDir, "ext"), 0755))
	phpIniPath := filepath.Join(versionDir, "php.ini")
	require.NoError(t, os.WriteFile(phpIniPath, []byte(ini), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "ext", "php_curl.dll"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "ext", "php_openssl.dll"), []byte(""), 0644))
	return phpIniPath
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
