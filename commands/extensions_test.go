package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		{name: "missing", kind: "error", message: "Extension missing not found in php.ini"},
	}, results)
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

func TestBuildExtensionInventory_GroupsConfiguredAndAvailableExtensions(t *testing.T) {
	ini := strings.Join([]string{
		`extension=php_curl.dll`,
		`;extension="ext\\php_openssl.dll"`,
		`extension=missing`,
	}, "\n")

	inventory := buildExtensionInventory(ini, []string{"php_curl.dll", "php_mbstring.dll", "php_openssl.dll"})

	assert.Equal(t, extensionInventory{
		Enabled:   []string{"curl", "missing (missing file)"},
		Disabled:  []string{"openssl"},
		Available: []string{"mbstring"},
	}, inventory)
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

func TestBuildExtensionInventory_PrefersEnabledStateWhenExtensionAppearsTwice(t *testing.T) {
	ini := strings.Join([]string{
		`;extension=php_curl.dll`,
		`extension=curl`,
	}, "\n")

	inventory := buildExtensionInventory(ini, []string{"php_curl.dll"})

	assert.Equal(t, []string{"curl"}, inventory.Enabled)
	assert.Empty(t, inventory.Disabled)
	assert.Empty(t, inventory.Available)
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
	require.NoError(t, os.MkdirAll(versionDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(versionDir, "ext"), 0755))
	phpIniPath := filepath.Join(versionDir, "php.ini")
	require.NoError(t, os.WriteFile(phpIniPath, []byte(ini), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "ext", "php_curl.dll"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "ext", "php_openssl.dll"), []byte(""), 0644))
	return phpIniPath
}
