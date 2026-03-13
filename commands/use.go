package commands

import (
	"fmt"
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Use(args []string) {
	threadSafe := true

	if len(args) < 1 {
		theme.Error("You must specify a version to use.")
		return
	}

	if len(args) > 1 {
		if args[1] == "nts" {
			threadSafe = false
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".pvm")); os.IsNotExist(err) {
		theme.Error("No PHP versions installed")
		return
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".pvm", "versions")); os.IsNotExist(err) {
		theme.Error("No PHP versions installed")
		return
	}

	binPath := filepath.Join(homeDir, ".pvm", "bin")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		os.Mkdir(binPath, 0755)
	}

	versions, err := os.ReadDir(filepath.Join(homeDir, ".pvm", "versions"))
	if err != nil {
		log.Fatalln(err)
	}

	var selectedVersion *versionMeta
	for i, version := range versions {
		safe := true
		if strings.Contains(version.Name(), "nts") || strings.Contains(version.Name(), "NTS") {
			safe = false
		}

		foundVersion := common.ComputeVersion(version.Name(), safe, "")
		if threadSafe == foundVersion.ThreadSafe && strings.HasPrefix(foundVersion.String(), args[0]) {
			selectedVersion = &versionMeta{
				number: foundVersion,
				folder: versions[i],
			}
		}
	}

	if selectedVersion == nil {
		theme.Error("The specified version is not installed.")
		return
	}

	requestedVersion := common.ComputeVersion(args[0], threadSafe, "")
	if requestedVersion.Minor == -1 {
		theme.Warning(fmt.Sprintf("No minor version specified, assumed newest minor version %s.", selectedVersion.number.String()))
	} else if requestedVersion.Patch == -1 {
		theme.Warning(fmt.Sprintf("No patch version specified, assumed newest patch version %s.", selectedVersion.number.String()))
	}

	batPath := filepath.Join(binPath, "php.bat")
	if _, err := os.Stat(batPath); err == nil {
		os.Remove(batPath)
	}

	shPath := filepath.Join(binPath, "php")
	if _, err := os.Stat(shPath); err == nil {
		os.Remove(shPath)
	}

	batPathCGI := filepath.Join(binPath, "php-cgi.bat")
	if _, err := os.Stat(batPathCGI); err == nil {
		os.Remove(batPathCGI)
	}

	shPathCGI := filepath.Join(binPath, "php-cgi")
	if _, err := os.Stat(shPathCGI); err == nil {
		os.Remove(shPathCGI)
	}

	batPathComposer := filepath.Join(binPath, "composer.bat")
	if _, err := os.Stat(batPathComposer); err == nil {
		os.Remove(batPathComposer)
	}

	shPathComposer := filepath.Join(binPath, "composer")
	if _, err := os.Stat(shPathComposer); err == nil {
		os.Remove(shPathComposer)
	}

	versionFolderPath := filepath.Join(homeDir, ".pvm", "versions", selectedVersion.folder.Name())
	versionPath := filepath.Join(versionFolderPath, "php.exe")
	versionPathCGI := filepath.Join(versionFolderPath, "php-cgi.exe")
	composerPath := filepath.Join(versionFolderPath, "composer", "composer.phar")

	batCommand := "@echo off \n"
	batCommand = batCommand + "set filepath=\"" + versionPath + "\"\n"
	batCommand = batCommand + "set arguments=%*\n"
	batCommand = batCommand + "%filepath% %arguments%\n"

	err = os.WriteFile(batPath, []byte(batCommand), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	shCommand := "#!/bin/bash\n"
	shCommand = shCommand + "filepath=\"" + versionPath + "\"\n"
	shCommand = shCommand + "\"$filepath\" \"$@\""

	err = os.WriteFile(shPath, []byte(shCommand), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	batCommandCGI := "@echo off \n"
	batCommandCGI = batCommandCGI + "set filepath=\"" + versionPathCGI + "\"\n"
	batCommandCGI = batCommandCGI + "set arguments=%*\n"
	batCommandCGI = batCommandCGI + "%filepath% %arguments%\n"

	err = os.WriteFile(batPathCGI, []byte(batCommandCGI), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	shCommandCGI := "#!/bin/bash\n"
	shCommandCGI = shCommandCGI + "filepath=\"" + versionPathCGI + "\"\n"
	shCommandCGI = shCommandCGI + "\"$filepath\" \"$@\""

	err = os.WriteFile(shPathCGI, []byte(shCommandCGI), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	batCommandComposer := "@echo off \n"
	batCommandComposer = batCommandComposer + "set filepath=\"" + versionPath + "\"\n"
	batCommandComposer = batCommandComposer + "set composerpath=\"" + composerPath + "\"\n"
	batCommandComposer = batCommandComposer + "set arguments=%*\n"
	batCommandComposer = batCommandComposer + "%filepath% %composerpath% %arguments%\n"

	err = os.WriteFile(batPathComposer, []byte(batCommandComposer), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	shCommandComposer := "#!/bin/bash\n"
	shCommandComposer = shCommandComposer + "filepath=\"" + versionPath + "\"\n"
	shCommandComposer = shCommandComposer + "composerpath=\"" + composerPath + "\"\n"
	shCommandComposer = shCommandComposer + "\"$filepath\" \"$composerpath\" \"$@\""

	err = os.WriteFile(shPathComposer, []byte(shCommandComposer), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	extensionDirPath := filepath.Join(versionFolderPath, "ext")
	extensionLinkPath := filepath.Join(binPath, "ext")

	if _, err := os.Stat(extensionLinkPath); err == nil {
		cmd := exec.Command("cmd", "/C", "rmdir", extensionLinkPath)
		_, err := cmd.Output()
		if err != nil {
			log.Fatalln("Error deleting ext directory directory link:", err)
			return
		}
	}

	cmd := exec.Command("cmd", "/C", "mklink", "/J", extensionLinkPath, extensionDirPath)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalln("Error creating ext directory symlink:", err)
		return
	}
	theme.Info(string(output))

	err = os.WriteFile(filepath.Join(homeDir, ".pvm", "version"), []byte(selectedVersion.folder.Name()), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	theme.Success(fmt.Sprintf("Using PHP %s", selectedVersion.number))
}

type versionMeta struct {
	number common.Version
	folder os.DirEntry
}
