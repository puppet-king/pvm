package common

import (
	"os"
	"path/filepath"
)

type PVMPaths struct {
	Home               string
	Root               string
	VersionsDir        string
	BinDir             string
	CurrentVersionFile string
}

func NewPVMPaths() (PVMPaths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return PVMPaths{}, err
	}

	root := filepath.Join(homeDir, ".pvm")
	return PVMPaths{
		Home:               homeDir,
		Root:               root,
		VersionsDir:        filepath.Join(root, "versions"),
		BinDir:             filepath.Join(root, "bin"),
		CurrentVersionFile: filepath.Join(root, "version"),
	}, nil
}

func (p PVMPaths) VersionDir(name string) string {
	return filepath.Join(p.VersionsDir, name)
}
