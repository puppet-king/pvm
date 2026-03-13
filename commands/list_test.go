package commands

import (
	"fmt"
	"strings"
	"testing"

	"hjbdev/pvm/common"

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

func TestListLocal_ReturnsErrorForInvalidCurrentVersionMetadata(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)
	writeInstalledVersion(t, homeDir, "8.3.1")
	writeActiveVersionMetadata(t, homeDir, "not-a-version")

	err := ListLocal()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid current version metadata")
}

func TestListRemote_MarksInstalledVersionsAndSortsOutput(t *testing.T) {
	previousRetrieveRemote := retrievePHPVersions
	previousRetrieveInstalled := retrieveInstalledPHPVersions
	t.Cleanup(func() {
		retrievePHPVersions = previousRetrieveRemote
		retrieveInstalledPHPVersions = previousRetrieveInstalled
	})

	retrievePHPVersions = func() ([]common.Version, error) {
		return []common.Version{
			{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true},
			{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true},
		}, nil
	}
	retrieveInstalledPHPVersions = func() ([]common.Version, error) {
		return []common.Version{{Major: 8, Minor: 3, Patch: 10, ThreadSafe: true}}, nil
	}

	output := captureStdout(t, func() {
		err := ListRemote()
		require.NoError(t, err)
	})

	assert.Contains(t, output, "PHP versions available")
	assert.Contains(t, output, "*   8.3.10")
	assert.Contains(t, output, "    8.4.1")
	assert.Less(t, indexOfLine(output, "*   8.3.10"), indexOfLine(output, "    8.4.1"))
}

func TestListRemote_ReturnsFetchError(t *testing.T) {
	previousRetrieveRemote := retrievePHPVersions
	t.Cleanup(func() {
		retrievePHPVersions = previousRetrieveRemote
	})

	retrievePHPVersions = func() ([]common.Version, error) {
		return nil, fmt.Errorf("boom")
	}

	err := ListRemote()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func indexOfLine(output string, line string) int {
	return strings.Index(output, line)
}
