package commands

import (
	"testing"

	"hjbdev/pvm/common"

	"github.com/stretchr/testify/assert"
)

func TestFindLatestMinor_PrefersNewestMinorBeforePatch(t *testing.T) {
	versions := []common.Version{
		{Major: 8, Minor: 3, Patch: 20, ThreadSafe: true},
		{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true},
		{Major: 8, Minor: 2, Patch: 99, ThreadSafe: true},
	}

	version := FindLatestMinor(versions, 8, true)

	assert.Equal(t, common.Version{Major: 8, Minor: 4, Patch: 1, ThreadSafe: true}, version)
}

func TestFindLatestMinor_PrefersLatestPatchWithinMinor(t *testing.T) {
	versions := []common.Version{
		{Major: 8, Minor: 4, Patch: 0, ThreadSafe: true},
		{Major: 8, Minor: 4, Patch: 2, ThreadSafe: true},
		{Major: 8, Minor: 3, Patch: 9, ThreadSafe: true},
	}

	version := FindLatestMinor(versions, 8, true)

	assert.Equal(t, common.Version{Major: 8, Minor: 4, Patch: 2, ThreadSafe: true}, version)
}
