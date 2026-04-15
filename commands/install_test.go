package commands

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"hjbdev/pvm/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindLatestMinor_PrefersNewestMinorBeforePatch(t *testing.T) {
	versions := []common.Version{
		{Major: 8, Minor: 3, Patch: 20, ThreadSafe: true},
		{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true},
		{Major: 8, Minor: 2, Patch: 99, ThreadSafe: true},
	}
	spec := common.VersionSpec{Major: 8, ThreadSafe: true}

	version, ok := findLatestMinor(versions, spec)

	assert.True(t, ok)
	assert.Equal(t, common.Version{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true}, version)
}

func TestFindLatestMinor_PrefersLatestPatchWithinMinor(t *testing.T) {
	versions := []common.Version{
		{Major: 8, Minor: 4, Patch: 0, ThreadSafe: true},
		{Major: 8, Minor: 4, Patch: 2, ThreadSafe: true},
		{Major: 8, Minor: 3, Patch: 9, ThreadSafe: true},
	}
	spec := common.VersionSpec{Major: 8, ThreadSafe: true}

	version, ok := findLatestMinor(versions, spec)

	assert.True(t, ok)
	assert.Equal(t, common.Version{Major: 8, Minor: 4, Patch: 2, ThreadSafe: true}, version)
}

func TestResolveInstallVersion_PrefersExactMatchForFullSpec(t *testing.T) {
	versions := []common.Version{
		{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true},
		{Major: 8, Minor: 3, Patch: 1, ThreadSafe: true},
	}
	spec := common.VersionSpec{Major: 8, Minor: 3, Patch: 1, HasMinor: true, HasPatch: true, ThreadSafe: true}

	version, ok := resolveInstallVersion(versions, spec)

	assert.True(t, ok)
	assert.Equal(t, common.Version{Major: 8, Minor: 3, Patch: 1, ThreadSafe: true}, version)
}

func TestDownloadPHPArchive_ReturnsErrorWhenArchiveAlreadyExists(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	paths, err := common.NewPVMPaths()
	require.NoError(t, err)
	require.NoError(t, ensureInstallDirs(paths))

	version := common.Version{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true, Url: "https://example.com/php-8.3.10-Win32-vs16-x64.zip"}
	archivePath := filepath.Join(paths.VersionsDir, "php-8.3.10-Win32-vs16-x64.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("existing"), 0644))

	_, _, err = downloadPHPArchive(paths, version)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestComposerURLForVersion_UsesLTSForPHP71(t *testing.T) {
	assert.Equal(t, "https://getcomposer.org/download/latest-2.2.x/composer.phar", composerURLForVersion(common.Version{Major: 7, Minor: 1, Patch: 33, ThreadSafe: true}))
}

func TestComposerURLForVersion_UsesStableForPHP72Plus(t *testing.T) {
	assert.Equal(t, "https://getcomposer.org/download/latest-stable/composer.phar", composerURLForVersion(common.Version{Major: 7, Minor: 2, Patch: 0, ThreadSafe: true}))
}

func TestDownloadFile_ReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	err := downloadFile(server.URL, filepath.Join(t.TempDir(), "download.zip"), 0)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestUnzip_ExtractsArchiveContents(t *testing.T) {
	src := filepath.Join(t.TempDir(), "archive.zip")
	createZipFile(t, src, map[string]string{
		"php.exe":          "php",
		"ext/php_curl.dll": "curl",
	})
	dest := filepath.Join(t.TempDir(), "php")

	err := unzip(src, dest)

	require.NoError(t, err)
	contents, err := os.ReadFile(filepath.Join(dest, "php.exe"))
	require.NoError(t, err)
	assert.Equal(t, "php", string(contents))
	contents, err = os.ReadFile(filepath.Join(dest, "ext", "php_curl.dll"))
	require.NoError(t, err)
	assert.Equal(t, "curl", string(contents))
}

func TestExtractZipFile_RejectsZipSlipEntry(t *testing.T) {
	src := filepath.Join(t.TempDir(), "archive.zip")
	createZipFile(t, src, map[string]string{
		"../evil.txt": "bad",
	})

	reader, err := zip.OpenReader(src)
	require.NoError(t, err)
	defer reader.Close()

	err = extractZipFile(filepath.Join(t.TempDir(), "dest"), reader.File[0])

	require.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path")
}

func createZipFile(t *testing.T, path string, files map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, contents := range files {
		entry, err := writer.Create(name)
		require.NoError(t, err)
		_, err = entry.Write([]byte(contents))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
}
