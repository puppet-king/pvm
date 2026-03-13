package commands

import (
	"fmt"
	"hjbdev/pvm/common"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type installedVersion struct {
	number common.Version
	folder os.DirEntry
}

type launcherSpec struct {
	name       string
	windowsCmd string
	unixCmd    string
}

func resolveInstalledVersion(entries []os.DirEntry, spec common.VersionSpec) (installedVersion, bool) {
	matches := make([]installedVersion, 0)
	for _, entry := range entries {
		version, err := common.ParseVersion(entry.Name(), common.IsThreadSafeName(entry.Name()), "")
		if err != nil {
			continue
		}
		if !spec.Matches(version) {
			continue
		}

		matches = append(matches, installedVersion{number: version, folder: entry})
	}

	if len(matches) == 0 {
		return installedVersion{}, false
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[j].number.LessThan(matches[i].number)
	})

	return matches[0], true
}

func selectionWarning(spec common.VersionSpec, version common.Version) string {
	if !spec.HasMinor {
		return fmt.Sprintf("No minor version specified, assumed newest minor version %s.", version.String())
	}
	if !spec.HasPatch {
		return fmt.Sprintf("No patch version specified, assumed newest patch version %s.", version.String())
	}

	return ""
}

func launcherSpecs(versionDir string) []launcherSpec {
	phpPath := filepath.Join(versionDir, "php.exe")
	phpCGIPath := filepath.Join(versionDir, "php-cgi.exe")
	composerPath := filepath.Join(versionDir, "composer", "composer.phar")

	return []launcherSpec{
		{
			name:       "php",
			windowsCmd: fmt.Sprintf("@echo off\r\nset filepath=\"%s\"\r\nset arguments=%%*\r\n%%filepath%% %%arguments%%\r\n", phpPath),
			unixCmd:    fmt.Sprintf("#!/bin/bash\nfilepath=\"%s\"\n\"$filepath\" \"$@\"", phpPath),
		},
		{
			name:       "php-cgi",
			windowsCmd: fmt.Sprintf("@echo off\r\nset filepath=\"%s\"\r\nset arguments=%%*\r\n%%filepath%% %%arguments%%\r\n", phpCGIPath),
			unixCmd:    fmt.Sprintf("#!/bin/bash\nfilepath=\"%s\"\n\"$filepath\" \"$@\"", phpCGIPath),
		},
		{
			name:       "composer",
			windowsCmd: fmt.Sprintf("@echo off\r\nset filepath=\"%s\"\r\nset composerpath=\"%s\"\r\nset arguments=%%*\r\n%%filepath%% %%composerpath%% %%arguments%%\r\n", phpPath, composerPath),
			unixCmd:    fmt.Sprintf("#!/bin/bash\nfilepath=\"%s\"\ncomposerpath=\"%s\"\n\"$filepath\" \"$composerpath\" \"$@\"", phpPath, composerPath),
		},
	}
}

func removeIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

func writeLaunchers(binDir string, specs []launcherSpec) error {
	for _, spec := range specs {
		windowsPath := filepath.Join(binDir, spec.name+".bat")
		unixPath := filepath.Join(binDir, spec.name)

		if err := removeIfExists(windowsPath); err != nil {
			return err
		}
		if err := removeIfExists(unixPath); err != nil {
			return err
		}
		if err := os.WriteFile(windowsPath, []byte(spec.windowsCmd), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(unixPath, []byte(spec.unixCmd), 0755); err != nil {
			return err
		}
	}

	return nil
}

func extLinkOutput(output []byte) string {
	return strings.TrimSpace(string(output))
}
