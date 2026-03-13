package commands

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBin_PrintsPvmBinPath(t *testing.T) {
	homeDir := t.TempDir()
	setHomeDir(t, homeDir)

	output := captureStdout(t, func() {
		require.NoError(t, Bin())
	})

	assert.Contains(t, output, "Add the following directory to your PATH:")
	assert.Contains(t, output, filepath.Join(homeDir, ".pvm", "bin"))
}
