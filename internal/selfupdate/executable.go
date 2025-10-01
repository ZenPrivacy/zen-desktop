package selfupdate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/constants"
)

func replaceExecutable(tempDir, destPath string) error {
	expectedExecName := constants.AppName
	if runtime.GOOS == "windows" {
		expectedExecName += ".exe"
	}
	src := filepath.Join(tempDir, expectedExecName)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("expected executable '%s' not found", expectedExecName)
	}

	tmpDst := destPath + ".part"

	if err := copyFile(src, tmpDst); err != nil {
		_ = os.Remove(tmpDst)
		return fmt.Errorf("copy new executable: %w", err)
	}

	if fi, err := os.Stat(destPath); err == nil {
		_ = os.Chmod(tmpDst, fi.Mode())
	} else {
		_ = os.Chmod(tmpDst, 0755)
	}

	if err := os.Rename(tmpDst, destPath); err != nil {
		_ = os.Remove(tmpDst)
		return fmt.Errorf("rename new executable: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir dst dir: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy: %w", err)
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return fmt.Errorf("sync: %w", err)
	}
	return out.Close()
}

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
