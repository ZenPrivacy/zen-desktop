package cfg

import (
	"os"
	"path/filepath"

	"github.com/ZenPrivacy/zen-desktop/internal/constants"
)

const (
	appFolderName = constants.AppName
	configDirName = "Config"
)

func getConfigDir() (string, error) {
	// According to Apple's guidelines, files in the ~/Library/Preferences should be only managed using native APIs,
	// so we use a subfolder in ~/Library/Application Support instead.
	// https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemOverview/FileSystemOverview.html#//apple_ref/doc/uid/TP40010672-CH2-SW1

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(homeDir, "Library", "Application Support", appFolderName, configDirName)
	return dir, nil
}

func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(homeDir, "Library", "Application Support", appFolderName)
	return dir, nil
}
