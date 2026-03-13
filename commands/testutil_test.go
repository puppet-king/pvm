package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
)

func setHomeDir(t *testing.T, homeDir string) {
	t.Helper()
	t.Setenv("HOME", homeDir)
}

func writeInstalledVersion(t *testing.T, homeDir string, version string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pvm", "versions", version), 0755))
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	previousStdout := os.Stdout
	previousColorOutput := color.Output
	previousNoColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() {
		os.Stdout = previousStdout
		color.Output = previousColorOutput
		color.NoColor = previousNoColor
	})

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	color.Output = writer

	outputCh := make(chan string, 1)
	go func() {
		var buffer bytes.Buffer
		_, _ = io.Copy(&buffer, reader)
		outputCh <- buffer.String()
	}()

	fn()
	require.NoError(t, writer.Close())

	output := <-outputCh
	require.NoError(t, reader.Close())
	return output
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
