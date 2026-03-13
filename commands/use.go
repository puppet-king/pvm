package commands

import (
	"fmt"
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Use(args []string) error {
	return runUse(args)
}

func runUse(args []string) error {
	request, err := parseUseRequest(args)
	if err != nil {
		return err
	}

	paths, err := common.NewPVMPaths()
	if err != nil {
		return err
	}
	if err := ensureVersionsDir(paths); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.BinDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(paths.VersionsDir)
	if err != nil {
		return err
	}

	selected, ok := resolveInstalledVersion(entries, request)
	if !ok {
		return fmt.Errorf("the specified version is not installed")
	}

	if warning := selectionWarning(request, selected.number); warning != "" {
		theme.Warning(warning)
	}

	versionDir := paths.VersionDir(selected.folder.Name())
	if err := writeLaunchers(paths.BinDir, launcherSpecs(versionDir)); err != nil {
		return err
	}
	if err := refreshExtLink(paths.BinDir, filepath.Join(versionDir, "ext")); err != nil {
		return err
	}
	if err := os.WriteFile(paths.CurrentVersionFile, []byte(selected.folder.Name()), 0644); err != nil {
		return err
	}

	theme.Success(fmt.Sprintf("Using PHP %s", selected.number))
	return nil
}

func parseUseRequest(args []string) (common.VersionSpec, error) {
	if len(args) == 0 {
		return common.VersionSpec{}, fmt.Errorf("you must specify a version to use")
	}

	threadSafe := true
	if len(args) > 1 && args[1] == "nts" {
		threadSafe = false
	}

	return common.ParseVersionSpec(args[0], threadSafe)
}

func ensureVersionsDir(paths common.PVMPaths) error {
	if _, err := os.Stat(paths.Root); os.IsNotExist(err) {
		return fmt.Errorf("no PHP versions installed")
	}
	if _, err := os.Stat(paths.VersionsDir); os.IsNotExist(err) {
		return fmt.Errorf("no PHP versions installed")
	}

	return nil
}

func refreshExtLink(binDir string, extensionDir string) error {
	linkPath := filepath.Join(binDir, "ext")
	if _, err := os.Lstat(linkPath); err == nil {
		if runtime.GOOS == "windows" {
			cmd := exec.Command("cmd", "/C", "rmdir", linkPath)
			if _, err := cmd.Output(); err != nil {
				return fmt.Errorf("error deleting ext directory link: %w", err)
			}
		} else if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("error deleting ext directory link: %w", err)
		}
	}

	if runtime.GOOS != "windows" {
		return os.Symlink(extensionDir, linkPath)
	}

	cmd := exec.Command("cmd", "/C", "mklink", "/J", linkPath, extensionDir)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error creating ext directory symlink: %w", err)
	}
	if message := extLinkOutput(output); message != "" {
		theme.Info(message)
	}

	return nil
}
