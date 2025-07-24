package selfupdate

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

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
		if err != nil{
			t.Fatalf("failed to write zip file: %v", err)
		}
		
		parentDir := t.TempDir()
		dest := filepath.Join(parentDir, "extract")
		
		err = unzip(tmpZip, dest)
		if err == nil {
			t.Fatal("expected error due to path traversal, got nil")
		}
		
		entries, err := os.ReadDir(parentDir)
		if err != nil{
			t.Fatalf("failed to read parent dir: %v", err)
		}
		
		for _, entry := range entries{
			if entry.Name() == "extract"{
				continue
			}
			t.Fatalf("unexpected file created outside destination: %s", entry.Name())
		}
	})
}
