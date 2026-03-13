package common

import (
	"os"
	"path/filepath"
	"strings"
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

func Test_NormalizeExtensionName_HandlesPathsQuotesAndDllPrefix(t *testing.T) {
	assert.Equal(t, "curl", NormalizeExtensionName(`"ext\\php_curl.dll" ; comment`))
	assert.Equal(t, "openssl", NormalizeExtensionName(`C:/php/ext/php_openssl.dll`))
	assert.Equal(t, "xdebug", NormalizeExtensionName(`xdebug`))
}

func Test_ParsePhpIniExtensions_ParsesNormalizedEntries(t *testing.T) {
	ini := strings.Join([]string{
		`extension=php_curl.dll`,
		`;extension="ext\\php_openssl.dll"`,
		`extension=C:/php/ext/php_mbstring.dll ; comment`,
		`zend_extension=php_xdebug.dll`,
	}, "\n")

	extensions := ParsePhpIniExtensions(ini)

	assert.Equal(t, []PhpIniExtension{
		{Name: "curl", Enabled: true, Line: 0, Kind: PhpIniExtensionDirective},
		{Name: "openssl", Enabled: false, Line: 1, Kind: PhpIniExtensionDirective},
		{Name: "mbstring", Enabled: true, Line: 2, Kind: PhpIniExtensionDirective},
		{Name: "xdebug", Enabled: true, Line: 3, Kind: PhpIniZendExtensionDirective},
	}, extensions)
}

func Test_ParsePhpIniExtensions_IgnoresCommentProse(t *testing.T) {
	ini := strings.Join([]string{
		`; <ext> is the name of the extension. Do not put extension=foo in this comment.`,
		`; zend_extension=/full/path/to/xdebug.dll is also documented here`,
		`extension=php_curl.dll`,
	}, "\n")

	extensions := ParsePhpIniExtensions(ini)

	assert.Equal(t, []PhpIniExtension{
		{Name: "curl", Enabled: true, Line: 2, Kind: PhpIniExtensionDirective},
	}, extensions)
}

func Test_GetExtensionStatus_IgnoresZendExtensions(t *testing.T) {
	ini := strings.Join([]string{
		`zend_extension=php_xdebug.dll`,
		`extension=php_curl.dll`,
	}, "\n")

	status, line := GetExtensionStatus(ini, "xdebug")
	assert.Equal(t, ExtensionNotFound, status)
	assert.Equal(t, -1, line)

	status, line = GetExtensionStatus(ini, "curl")
	assert.Equal(t, ExtensionEnabled, status)
	assert.Equal(t, 1, line)
}

func Test_GetDirectiveStatus_FindsZendAndRegularExtensions(t *testing.T) {
	ini := strings.Join([]string{
		`;zend_extension=php_xdebug.dll`,
		`extension=php_curl.dll`,
	}, "\n")

	status, line, kind := GetDirectiveStatus(ini, "xdebug")
	assert.Equal(t, ExtensionDisabled, status)
	assert.Equal(t, 0, line)
	assert.Equal(t, PhpIniZendExtensionDirective, kind)

	status, line, kind = GetDirectiveStatus(ini, "curl")
	assert.Equal(t, ExtensionEnabled, status)
	assert.Equal(t, 1, line)
	assert.Equal(t, PhpIniExtensionDirective, kind)
}
