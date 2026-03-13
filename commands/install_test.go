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
