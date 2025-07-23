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
	// use LOCALAPPDATA instead of APPDATA because the config includes some machine-specific data
	// e.g. whether the CA has been installed
	if os.Getenv("LOCALAPPDATA") != "" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), appFolderName, configDirName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, "AppData", "Local", appFolderName, configDirName), nil
}

func getDataDir() (string, error) {
	if os.Getenv("LOCALAPPDATA") != "" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), appFolderName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, "AppData", "Local", appFolderName), nil
}
