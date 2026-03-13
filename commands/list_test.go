package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_MarksCurrentVersionFromVersionMetadata(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	writeInstalledVersion(t, homeDir, "8.2.15")
	writeInstalledVersion(t, homeDir, "8.3.1")
	writeActiveVersionMetadata(t, homeDir, "8.3.1")

	output := captureStdout(t, func() {
		List(nil)
	})

	assert.Contains(t, output, "8.3.1 (current)")
	assert.NotContains(t, output, "8.2.15 (current)")
}

func TestList_SkipsCurrentMarkerWithoutVersionMetadata(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	writeInstalledVersion(t, homeDir, "8.3.1")

	output := captureStdout(t, func() {
		List(nil)
	})

	assert.Contains(t, output, "8.3.1")
	assert.NotContains(t, output, "(current)")
}

func TestParseListAction_DefaultsToLocal(t *testing.T) {
	action, ok := parseListAction(nil)

	assert.True(t, ok)
	assert.Equal(t, "local", action)
}

func TestParseListAction_AcceptsRemote(t *testing.T) {
	action, ok := parseListAction([]string{"remote"})

	assert.True(t, ok)
	assert.Equal(t, "remote", action)
}

func TestParseListAction_RejectsUnknownSubcommands(t *testing.T) {
	action, ok := parseListAction([]string{"weird"})

	assert.False(t, ok)
	assert.Equal(t, "", action)
}

func TestList_ShowsUsageForInvalidSubcommands(t *testing.T) {
	output := captureStdout(t, func() {
		List([]string{"weird"})
	})

	assert.Contains(t, output, "Invalid list action.")
	assert.Contains(t, output, "Usage: pvm list [remote]")
}

func writeInstalledVersion(t *testing.T, homeDir string, version string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pvm", "versions", version), 0755))
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
