package commands

import (
	"archive/zip"
	"fmt"
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

func Install(args []string) error {
	return runInstall(args)
}

func runInstall(args []string) error {
	request, err := parseInstallRequest(args)
	if err != nil {
		return err
	}

	threadSafetyLabel := "thread safe"
	if !request.ThreadSafe {
		threadSafetyLabel = "non-thread safe"
	}
	theme.Warning(fmt.Sprintf("%s version will be installed", titlePhrase(threadSafetyLabel)))

	versions, err := common.RetrievePHPVersions()
	if err != nil {
		return err
	}

	desiredVersion, ok := resolveInstallVersion(versions, request)
	if !ok {
		return fmt.Errorf("could not find the desired version: %s %s", args[1], threadSafetyLabel)
	}

	fmt.Printf("Installing PHP %s\n", desiredVersion)

	paths, err := common.NewPVMPaths()
	if err != nil {
		return err
	}
	if err := ensureInstallDirs(paths); err != nil {
		return err
	}

	archivePath, phpPath, err := downloadPHPArchive(paths, desiredVersion)
	if err != nil {
		return err
	}
	if err := unzip(archivePath, phpPath); err != nil {
		return err
	}
	theme.Info("Cleaning up")
	if err := os.Remove(archivePath); err != nil {
		return err
	}
	if err := installComposerForVersion(phpPath, desiredVersion); err != nil {
		return err
	}

	theme.Success(fmt.Sprintf("Finished installing PHP %s", desiredVersion))
	return nil
}

func parseInstallRequest(args []string) (common.VersionSpec, error) {
	if len(args) < 2 {
		return common.VersionSpec{}, fmt.Errorf("you must specify a version to install")
	}

	threadSafe := true
	if len(args) > 2 && args[2] == "nts" {
		threadSafe = false
	}

	return common.ParseVersionSpec(args[1], threadSafe)
}

func ensureInstallDirs(paths common.PVMPaths) error {
	if err := os.MkdirAll(paths.VersionsDir, 0755); err != nil {
		return err
	}

	return nil
}

func downloadPHPArchive(paths common.PVMPaths, version common.Version) (string, string, error) {
	theme.Info("Downloading")
	zipFileName := path.Base(version.Url)
	zipPath := filepath.Join(paths.VersionsDir, zipFileName)
	if _, err := os.Stat(zipPath); err == nil {
		return "", "", fmt.Errorf("PHP %s already exists", version)
	}
	if err := downloadFile(version.Url, zipPath, version.SizeBytes); err != nil {
		return "", "", fmt.Errorf("error while downloading PHP from %s: %w", version.Url, err)
	}

	phpFolder := strings.TrimSuffix(zipFileName, ".zip")
	return zipPath, filepath.Join(paths.VersionsDir, phpFolder), nil
}

func installComposerForVersion(phpPath string, version common.Version) error {
	composerFolderPath := filepath.Join(phpPath, "composer")
	if err := os.MkdirAll(composerFolderPath, 0755); err != nil {
		return err
	}

	composerURL := composerURLForVersion(version)
	composerPath := filepath.Join(composerFolderPath, "composer.phar")
	if err := downloadFile(composerURL, composerPath, 0); err != nil {
		return fmt.Errorf("error while downloading Composer from %v: %w", composerURL, err)
	}

	return nil
}

func composerURLForVersion(version common.Version) string {
	if version.Compare(common.Version{Major: 7, Minor: 2, Patch: 0}) == -1 {
		return "https://getcomposer.org/download/latest-2.2.x/composer.phar"
	}

	return "https://getcomposer.org/download/latest-stable/composer.phar"
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, file := range r.File {
		if err := extractZipFile(dest, file); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(dest string, file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	targetPath := filepath.Join(dest, file.Name)
	if !strings.HasPrefix(targetPath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", targetPath)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, file.Mode())
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func downloadFile(fileURL string, filePath string, sizeBytes int64) error {
	client := common.NewHTTPClient()

	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("HTTP_PROXY") != "" {
		// 只要这两个变量有一个不为空，就提示正在使用系统代理
		theme.Info("Using system proxy configuration")
	}

	// Now download the file
	response, err := client.Get(fileURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while downloading %s: %d", fileURL, response.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Use provided sizeBytes if available, otherwise try response.ContentLength
	contentLength := sizeBytes
	if contentLength <= 0 {
		contentLength = response.ContentLength
	}

	if contentLength > 0 {
		bar := pb.Full.Start64(contentLength)
		writer := bar.NewProxyWriter(out)
		_, err = io.Copy(writer, response.Body)
		bar.Finish()
		return err
	}

	_, err = io.Copy(out, response.Body)
	return err
}

func titlePhrase(value string) string {
	if value == "" {
		return ""
	}

	return strings.ToUpper(value[:1]) + value[1:]
}
