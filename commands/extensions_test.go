package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeExtensions_TrimsSkipsEmptyAndDeduplicates(t *testing.T) {
	extensions := normalizeExtensions("curl, mbstring ,, curl,openssl ")

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
	phpIniPath := filepath.Join(versionDir, "php.ini")
	require.NoError(t, os.WriteFile(phpIniPath, []byte(ini), 0644))
	return phpIniPath
}
