package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ParsePHPVersions_ParsesDownloadsPHPNetArchiveListing(t *testing.T) {
	body := `<html><body>
		<a href="/~windows/releases/">Parent Directory</a>
		<a href="php-8.3.14-Win32-vs16-x64.zip">php-8.3.14-Win32-vs16-x64.zip</a>
		<a href="php-8.3.14-nts-Win32-vs16-x64.zip">php-8.3.14-nts-Win32-vs16-x64.zip</a>
		<a href="php-debug-pack-8.3.14-Win32-vs16-x64.zip">php-debug-pack-8.3.14-Win32-vs16-x64.zip</a>
		<a href="php-8.3.14-Win32-vs16-x86.zip">php-8.3.14-Win32-vs16-x86.zip</a>
	</body></html>`

	versions, err := ParsePHPVersions(body, "https://downloads.php.net/~windows/releases/archives/")

	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, Version{
		Major:      8,
		Minor:      3,
		Patch:      14,
		ThreadSafe: true,
		Url:        "https://downloads.php.net/~windows/releases/archives/php-8.3.14-Win32-vs16-x64.zip",
	}, versions[0])
	assert.Equal(t, Version{
		Major:      8,
		Minor:      3,
		Patch:      14,
		ThreadSafe: false,
		Url:        "https://downloads.php.net/~windows/releases/archives/php-8.3.14-nts-Win32-vs16-x64.zip",
	}, versions[1])
}

func Test_ParsePHPVersions_ParsesLegacyUppercaseListing(t *testing.T) {
	body := `<A HREF="/downloads/releases/archives/php-8.4.1-Win32-vs17-x64.zip">php-8.4.1-Win32-vs17-x64.zip</A>`

	versions, err := ParsePHPVersions(body, "https://windows.php.net/downloads/releases/archives/")

	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "https://windows.php.net/downloads/releases/archives/php-8.4.1-Win32-vs17-x64.zip", versions[0].Url)
}

func Test_Version_Compare(t *testing.T) {
	v1 := Version{Major: 1, Minor: 2, Patch: 3, ThreadSafe: false}
	v2 := Version{Major: 1, Minor: 2, Patch: 4}
	v3 := Version{Major: 1, Minor: 2, Patch: 3, ThreadSafe: true}

	assert.Equal(t, v1.LessThan(v2), true)
	assert.Equal(t, v1.Same(v3), false)

	// testing versions with "nulled" (-1) values

	v4 := Version{Major: 1, Minor: 2, Patch: -1}
	v5 := Version{Major: 1, Minor: 2, Patch: 3}

	assert.Equal(t, v4.LessThan(v5), false)

	v6 := Version{Major: 1, Minor: -1}
	v7 := Version{Major: 1, Minor: 2}

	assert.Equal(t, v6.LessThan(v7), false)
}

func Test_ReadPhpIni_ReturnsFileContents(t *testing.T) {
	tempDir := t.TempDir()
	phpIniPath := filepath.Join(tempDir, "php.ini")
	require.NoError(t, os.WriteFile(phpIniPath, []byte("extension=curl"), 0644))

	contents, err := ReadPhpIni(phpIniPath)

	require.NoError(t, err)
	assert.Equal(t, "extension=curl", contents)
}

func Test_ReadPhpIni_ReturnsErrorForMissingFile(t *testing.T) {
	contents, err := ReadPhpIni(filepath.Join(t.TempDir(), "missing.ini"))

	assert.Error(t, err)
	assert.Equal(t, "", contents)
}

func Test_GetExtensionStatus_DetectsEnabledAndDisabledExtensions(t *testing.T) {
	ini := "extension=curl\n;extension=openssl\n"

	status, line := GetExtensionStatus(ini, "curl")
	assert.Equal(t, ExtensionEnabled, status)
	assert.Equal(t, 0, line)

	status, line = GetExtensionStatus(ini, "openssl")
	assert.Equal(t, ExtensionDisabled, status)
	assert.Equal(t, 1, line)

	status, line = GetExtensionStatus(ini, "mbstring")
	assert.Equal(t, ExtensionNotFound, status)
	assert.Equal(t, -1, line)
}
