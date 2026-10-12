package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/irbis-sh/zen-desktop/internal/constants"
)

type failingHTTPClient struct{}

func (f *failingHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated download failure")
}

func TestIsNewer(t *testing.T) {
	t.Parallel()

	t.Run("fails on invalid current version", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v100",
		}
		if _, err := su.isNewer("v1.0.0"); err == nil {
			t.Error("got nil, want error")
		}
	})

	t.Run("fails on invalid new version", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v1.0.0",
		}
		if _, err := su.isNewer("v100"); err == nil {
			t.Error("got nil, want error")
		}
	})

	t.Run("returns true if new version has larger patch version than current", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v3.2.2",
		}
		newer, err := su.isNewer("v3.2.3")
		if err != nil {
			t.Error("got error, want nil")
		}
		if !newer {
			t.Error("got false, want true")
		}
	})

	t.Run("returns true if new version has larger minor version than current", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v0.1.0",
		}
		newer, err := su.isNewer("v0.5.0")
		if err != nil {
			t.Error("got error, want nil")
		}
		if !newer {
			t.Error("got false, want true")
		}
	})

	t.Run("returns true if new version has larger major version than current", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v10.0.0",
		}
		newer, err := su.isNewer("v14.2.8")
		if err != nil {
			t.Error("got error, want nil")
		}
		if !newer {
			t.Error("got false, want true")
		}
	})

	t.Run("returns false if versions are equal", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v4.1.1",
		}
		newer, err := su.isNewer("v4.1.1")
		if err != nil {
			t.Error("got error, want nil")
		}
		if newer {
			t.Error("got true, want false")
		}
	})

	t.Run("returns false if new version is older than current", func(t *testing.T) {
		t.Parallel()
		su := SelfUpdater{
			version: "v1.0.0",
		}
		newer, err := su.isNewer("v0.9.9")
		if err != nil {
			t.Error("got error, want nil")
		}
		if newer {
			t.Error("got true, want false")
		}
	})
}

func TestDownloadAndVerifyFileCleansUpOnError(t *testing.T) {
	su := &SelfUpdater{
		httpClient: &failingHTTPClient{},
	}

	tmpDir := os.TempDir()
	before, _ := filepath.Glob(filepath.Join(tmpDir, "downloaded-*"))
	beforeSet := make(map[string]bool, len(before))
	for _, f := range before {
		beforeSet[f] = true
	}

	_, err := su.downloadAndVerifyFile("http://example.com/file.zip", "deadbeef")
	if err == nil {
		t.Fatal("got nil, want error")
	}

	after, _ := filepath.Glob(filepath.Join(tmpDir, "downloaded-*"))
	for _, f := range after {
		if !beforeSet[f] {
			os.Remove(f)
			t.Errorf("temp file was not cleaned up: %s", f)
		}
	}
}

func TestPathTraversal(t *testing.T) {

	t.Run("blocks traversal attack", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)

		f, err := zw.Create("../../evil.txt")
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}

		_, err = f.Write([]byte("should not escape"))
		if err != nil {
			t.Fatalf("failed to write to zip: %v", err)
		}

		if err := zw.Close(); err != nil {
			t.Fatalf("failed to close zip writer: %v", err)
		}

		tmpZip := filepath.Join(t.TempDir(), "malicious.zip")
		err = os.WriteFile(tmpZip, buf.Bytes(), 0644)
		if err != nil {
			t.Fatalf("failed to write zip file: %v", err)
		}

		parentDir := t.TempDir()
		dest := filepath.Join(parentDir, "extract")

		err = unzip(tmpZip, dest)
		if err == nil {
			t.Fatal("expected error due to path traversal, got nil")
		}

		entries, err := os.ReadDir(parentDir)
		if err != nil {
			t.Fatalf("failed to read parent dir: %v", err)
		}

		for _, entry := range entries {
			if entry.Name() == "extract" {
				continue
			}
			t.Fatalf("unexpected file created outside destination: %s", entry.Name())
		}
	})
}

func TestInstallExecutable(t *testing.T) {
	execName := constants.AppName
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}

	t.Run("replaces executable", func(t *testing.T) {
		execDir := t.TempDir()
		execPath := filepath.Join(execDir, execName)
		if err := os.WriteFile(execPath, []byte("old"), 0600); err != nil {
			t.Fatalf("write current executable: %v", err)
		}
		archivePath := writeTarGz(t, execName)
		setMissingTempDir(t)

		if err := installExecutable(archivePath, execPath); err != nil {
			t.Fatalf("installExecutable: %v", err)
		}

		got, err := os.ReadFile(execPath)
		if err != nil {
			t.Fatalf("read executable: %v", err)
		}
		if string(got) != "new" {
			t.Errorf("executable content: got %q, want %q", got, "new")
		}
		assertDirEntries(t, execDir, execName)
	})

	t.Run("removes stale staging dirs", func(t *testing.T) {
		execDir := t.TempDir()
		execPath := filepath.Join(execDir, execName)
		if err := os.WriteFile(execPath, []byte("old"), 0600); err != nil {
			t.Fatalf("write current executable: %v", err)
		}
		staleDir := filepath.Join(execDir, ".zen-update-123")
		if err := os.Mkdir(staleDir, 0700); err != nil {
			t.Fatalf("create stale staging dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(staleDir, execName), []byte("partial"), 0600); err != nil {
			t.Fatalf("write stale executable: %v", err)
		}
		archivePath := writeTarGz(t, execName)
		setMissingTempDir(t)

		if err := installExecutable(archivePath, execPath); err != nil {
			t.Fatalf("installExecutable: %v", err)
		}

		assertDirEntries(t, execDir, execName)
	})

	t.Run("replaces executable when exec dir is not a valid glob pattern", func(t *testing.T) {
		execDir := filepath.Join(t.TempDir(), "zen[")
		if err := os.Mkdir(execDir, 0700); err != nil {
			t.Fatalf("create exec dir: %v", err)
		}
		execPath := filepath.Join(execDir, execName)
		if err := os.WriteFile(execPath, []byte("old"), 0600); err != nil {
			t.Fatalf("write current executable: %v", err)
		}
		archivePath := writeTarGz(t, execName)
		setMissingTempDir(t)

		if err := installExecutable(archivePath, execPath); err != nil {
			t.Fatalf("installExecutable: %v", err)
		}

		got, err := os.ReadFile(execPath)
		if err != nil {
			t.Fatalf("read executable: %v", err)
		}
		if string(got) != "new" {
			t.Errorf("executable content: got %q, want %q", got, "new")
		}
	})

	t.Run("restores executable on failure", func(t *testing.T) {
		execDir := t.TempDir()
		execPath := filepath.Join(execDir, execName)
		if err := os.WriteFile(execPath, []byte("old"), 0600); err != nil {
			t.Fatalf("write current executable: %v", err)
		}
		archivePath := writeTarGz(t, "not-"+execName)
		setMissingTempDir(t)

		if err := installExecutable(archivePath, execPath); err == nil {
			t.Fatal("got nil, want error")
		}

		got, err := os.ReadFile(execPath)
		if err != nil {
			t.Fatalf("read executable: %v", err)
		}
		if string(got) != "old" {
			t.Errorf("executable content: got %q, want %q", got, "old")
		}
		assertDirEntries(t, execDir, execName)
	})
}

// setMissingTempDir points the system temp dir at a nonexistent directory so that
// staging the update there fails. The system temp dir may be on a different
// filesystem than the executable, in which case moving the new executable into
// place would fail.
func setMissingTempDir(t *testing.T) {
	t.Helper()

	missingDir := filepath.Join(t.TempDir(), "missing")
	t.Setenv("TMPDIR", missingDir)
	t.Setenv("TMP", missingDir)
	t.Setenv("TEMP", missingDir)
}

// writeTarGz writes a .tar.gz archive containing a single file with the given name
// and the content "new", and returns its path.
func writeTarGz(t *testing.T, name string) string {
	t.Helper()

	content := "new"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name:     name,
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	archivePath := filepath.Join(t.TempDir(), "update.tar.gz")
	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	return archivePath
}

func assertDirEntries(t *testing.T, dir string, want ...string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	if !slices.Equal(got, want) {
		t.Errorf("dir entries: got %v, want %v", got, want)
	}
}
