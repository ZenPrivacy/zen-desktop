package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func getExecPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	// https://github.com/golang/go/issues/40966
	if runtime.GOOS != "windows" {
		if execPath, err = filepath.EvalSymlinks(execPath); err != nil {
			return "", fmt.Errorf("eval symlinks: %w", err)
		}
	}

	return execPath, nil
}

func findAppBundlePath(execPath string) string {
	dir := filepath.Dir(execPath)
	for dir != "/" {
		if strings.HasSuffix(dir, ".app") {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
