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

func Test_ParsePHPVersions_ParsesSizeFromHTMLTable(t *testing.T) {
	body := `<html><body>
		<table>
		<tr><td><img src="icon.gif"></td><td><a href="php-8.3.14-Win32-vs16-x64.zip">php-8.3.14-Win32-vs16-x64.zip</a></td><td>2024-01-15 12:00</td><td>25M</td></tr>
		<tr><td><img src="icon.gif"></td><td><a href="php-8.3.14-nts-Win32-vs16-x64.zip">php-8.3.14-nts-Win32-vs16-x64.zip</a></td><td>2024-01-15 12:00</td><td>24M</td></tr>
		</table>
	</body></html>`

	versions, err := ParsePHPVersions(body, "https://downloads.php.net/~windows/releases/archives/")

	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, int64(25*1024*1024), versions[0].SizeBytes)
	assert.Equal(t, int64(24*1024*1024), versions[1].SizeBytes)
}

func Test_ParseSize_ParsesHumanReadableSizes(t *testing.T) {
	assert.Equal(t, int64(16*1024*1024), parseSize("16M"))
	assert.Equal(t, int64(25*1024*1024), parseSize("25M"))
	assert.Equal(t, int64(1024), parseSize("1K"))
	assert.Equal(t, int64(1024*1024*1024), parseSize("1G"))
	assert.Equal(t, int64(0), parseSize(""))
	assert.Equal(t, int64(0), parseSize("-"))
}

func Test_Version_Compare(t *testing.T) {
	v1 := Version{Major: 1, Minor: 2, Patch: 3, ThreadSafe: false}
	v2 := Version{Major: 1, Minor: 2, Patch: 4}
	v3 := Version{Major: 1, Minor: 2, Patch: 3, ThreadSafe: true}

	assert.Equal(t, v1.LessThan(v2), true)
	assert.Equal(t, v1.Same(v3), false)
}

func Test_ParseVersionSpec_AcceptsMajorMinorAndPatch(t *testing.T) {
	spec, err := ParseVersionSpec("8.3.10", true)

	assert.NoError(t, err)
	assert.Equal(t, VersionSpec{Major: 8, Minor: 3, Patch: 10, HasMinor: true, HasPatch: true, ThreadSafe: true}, spec)
}

func Test_ParseVersionSpec_AcceptsPartialVersions(t *testing.T) {
	spec, err := ParseVersionSpec("8.3", false)

	assert.NoError(t, err)
	assert.Equal(t, VersionSpec{Major: 8, Minor: 3, HasMinor: true, ThreadSafe: false}, spec)
}

func Test_ParseVersion_RequiresFullSemanticVersion(t *testing.T) {
	_, err := ParseVersion("8.3", true, "")

	assert.Error(t, err)
}

func Test_SortVersions_OrdersByVersionThenThreadSafety(t *testing.T) {
	versions := []Version{
		{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true},
		{Major: 8, Minor: 3, Patch: 10, ThreadSafe: false},
		{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true},
	}

	SortVersions(versions)

	assert.Equal(t, []Version{
		{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true},
		{Major: 8, Minor: 3, Patch: 10, ThreadSafe: false},
		{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true},
	}, versions)
}

func Test_IsThreadSafeName_IsCaseInsensitive(t *testing.T) {
	assert.True(t, IsThreadSafeName("php-8.3.10-Win32-vs16-x64.zip"))
	assert.False(t, IsThreadSafeName("php-8.3.10-NTS-Win32-vs16-x64.zip"))
}

func Test_IsSupportedArchiveName_RejectsUnsupportedArchives(t *testing.T) {
	assert.False(t, isSupportedArchiveName("php-debug-pack-8.3.14-Win32-vs16-x64.zip"))
	assert.False(t, isSupportedArchiveName("php-8.3.14-src.zip"))
	assert.False(t, isSupportedArchiveName("php-8.3.14-Win32-vs16-x86.zip"))
	assert.True(t, isSupportedArchiveName("php-8.3.14-Win32-vs16-x64.zip"))
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

func Test_ParsePhpIniExtensions_IgnoresBogusCommentExamplesInRealIni(t *testing.T) {
	ini := strings.Join([]string{
		`; Notes for Windows environments :`,
		`;   extension=mysqli`,
		`;   extension=/path/to/extension/mysqli.so`,
		`; Note : The syntax used in previous PHP versions ('extension=<ext>.so' and`,
		`; 'extension='php_<ext>.dll') is supported for legacy reasons and may be`,
		`; move to the new ('extension=<ext>) syntax.`,
		`;extension=curl`,
		`;extension=mbstring`,
		`;extension=openssl`,
		`; bogus example: extension='php_fake.dll') is supported for legacy reasons`,
		`; bogus example: zend_extension=/tmp/not-real-xdebug.dll is documented here`,
	}, "\n")

	extensions := ParsePhpIniExtensions(ini)
	names := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		names = append(names, extension.Name)
	}

	assert.NotContains(t, names, "fake")
	assert.NotContains(t, names, "not-real-xdebug")
	assert.NotContains(t, names, "php_<ext>")
	assert.NotContains(t, names, "ext>) syntax")
	assert.Contains(t, names, "curl")
	assert.Contains(t, names, "openssl")
	assert.Contains(t, names, "mbstring")
}
