package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelp_ShowsUpdatedCommandsAndAliases(t *testing.T) {
	output := captureStdout(t, func() {
		Help(false)
	})

	assert.Contains(t, output, "    list [remote] (alias: ls)")
	assert.Contains(t, output, "    bin")
	assert.Contains(t, output, "    install (alias: i)")
	assert.Contains(t, output, "    use <version> (alias: u)")
	assert.NotContains(t, output, "path")
	assert.NotContains(t, output, "list-remote")
	assert.NotContains(t, output, "ls-remote")
}
