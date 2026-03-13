package commands

import (
	"os"
	"path/filepath"
	"testing"

	"hjbdev/pvm/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInstalledVersion_PrefersExactPatchOverPrefixCollision(t *testing.T) {
	spec, err := common.ParseVersionSpec("8.3.1", true)
	require.NoError(t, err)

	entries := []os.DirEntry{
		mustDirEntry(t, "8.3.10"),
		mustDirEntry(t, "8.3.1"),
	}

	selected, ok := resolveInstalledVersion(entries, spec)

	assert.True(t, ok)
	assert.Equal(t, "8.3.1", selected.folder.Name())
}

func TestResolveInstalledVersion_SelectsLatestPatchForMinorRequest(t *testing.T) {
	spec, err := common.ParseVersionSpec("8.3", true)
	require.NoError(t, err)

	entries := []os.DirEntry{
		mustDirEntry(t, "8.3.1"),
		mustDirEntry(t, "8.3.10"),
		mustDirEntry(t, "8.2.20"),
	}

	selected, ok := resolveInstalledVersion(entries, spec)

	assert.True(t, ok)
	assert.Equal(t, "8.3.10", selected.folder.Name())
}

func TestResolveInstalledVersion_SelectsLatestMinorForMajorRequest(t *testing.T) {
	spec, err := common.ParseVersionSpec("8", true)
	require.NoError(t, err)

	entries := []os.DirEntry{
		mustDirEntry(t, "8.2.20"),
		mustDirEntry(t, "8.4.1"),
		mustDirEntry(t, "8.3.10"),
	}

	selected, ok := resolveInstalledVersion(entries, spec)

	assert.True(t, ok)
	assert.Equal(t, "8.4.1", selected.folder.Name())
}

func TestResolveInstalledVersion_RespectsNTSSelection(t *testing.T) {
	spec, err := common.ParseVersionSpec("8.3", false)
	require.NoError(t, err)

	entries := []os.DirEntry{
		mustDirEntry(t, "8.3.10"),
		mustDirEntry(t, "8.3.10 nts"),
	}

	selected, ok := resolveInstalledVersion(entries, spec)

	assert.True(t, ok)
	assert.Equal(t, "8.3.10 nts", selected.folder.Name())
	assert.False(t, selected.number.ThreadSafe)
}

func TestUse_WritesLaunchersAndCurrentVersionMetadata(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	versionDir := filepath.Join(homeDir, ".pvm", "versions", "8.3.10")
	require.NoError(t, os.MkdirAll(filepath.Join(versionDir, "composer"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "php-cgi.exe"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "composer", "composer.phar"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(versionDir, "ext"), 0755))

	output := captureStdout(t, func() {
		Use([]string{"8.3"})
	})

	metadata, err := os.ReadFile(filepath.Join(homeDir, ".pvm", "version"))
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", string(metadata))
	assert.Contains(t, output, "No patch version specified")
	assert.Contains(t, output, "Using PHP 8.3.10 thread safe")

	phpBat, err := os.ReadFile(filepath.Join(homeDir, ".pvm", "bin", "php.bat"))
	require.NoError(t, err)
	assert.Contains(t, string(phpBat), filepath.Join(versionDir, "php.exe"))

	composerSh, err := os.ReadFile(filepath.Join(homeDir, ".pvm", "bin", "composer"))
	require.NoError(t, err)
	assert.Contains(t, string(composerSh), filepath.Join(versionDir, "composer", "composer.phar"))
}

func TestRunUse_ReturnsErrorWhenNoVersionsInstalled(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)

	err := runUse([]string{"8.3"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no PHP versions installed")
}

func TestWriteLaunchers_ReplacesExistingLauncherFiles(t *testing.T) {
	binDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "php.bat"), []byte("old"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "php"), []byte("old"), 0644))

	err := writeLaunchers(binDir, []launcherSpec{{
		name:       "php",
		windowsCmd: "new bat",
		unixCmd:    "new sh",
	}})

	require.NoError(t, err)
	batContents, err := os.ReadFile(filepath.Join(binDir, "php.bat"))
	require.NoError(t, err)
	assert.Equal(t, "new bat", string(batContents))
	shContents, err := os.ReadFile(filepath.Join(binDir, "php"))
	require.NoError(t, err)
	assert.Equal(t, "new sh", string(shContents))
}

func TestRefreshExtLink_ReplacesExistingLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem link test in short mode")
	}

	binDir := t.TempDir()
	firstTarget := filepath.Join(t.TempDir(), "ext-a")
	secondTarget := filepath.Join(t.TempDir(), "ext-b")
	require.NoError(t, os.MkdirAll(firstTarget, 0755))
	require.NoError(t, os.MkdirAll(secondTarget, 0755))
	require.NoError(t, os.Symlink(firstTarget, filepath.Join(binDir, "ext")))

	err := refreshExtLink(binDir, secondTarget)

	require.NoError(t, err)
	resolved, err := os.Readlink(filepath.Join(binDir, "ext"))
	require.NoError(t, err)
	assert.Equal(t, secondTarget, resolved)
}

func mustDirEntry(t *testing.T, name string) os.DirEntry {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(path, 0755))
	entry, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entry, 1)
	return entry[0]
}
