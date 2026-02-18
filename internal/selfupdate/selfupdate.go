package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ZenPrivacy/zen-desktop/internal/cfg"
)

// NoSelfUpdate is set to "true" for builds distributed to package managers to prevent auto-updating. It is typed as a string because the linker allows only setting string variables at compile time (see https://pkg.go.dev/cmd/link).
// Set at compile time using ldflags (see the prod-noupdate task in the /tasks/build directory).
var NoSelfUpdate = "false"

const (
	// releaseTrack is the release track to follow for updates. It currently only takes the value "stable".
	releaseTrack = "stable"
	// manifestsBaseURL is the base URL for fetching update manifests.
	manifestsBaseURL = "https://update-manifests.zenprivacy.net"
	// checkInterval sets the frequency with which to check for updates.
	checkInterval = 6 * time.Hour
	// httpTimeout is the timeout for HTTP requests when checking for updates and downloading update files.
	httpTimeout = 5 * time.Minute
)

type release struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	AssetURL    string `json:"assetURL"`
	SHA256      string `json:"sha256"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type eventsEmitter interface {
	OnUpdateAvailable()
}

type SelfUpdater struct {
	version       string
	config        *cfg.Config
	httpClient    httpClient
	eventsEmitter eventsEmitter

	// execUpdateMu ensures that only one applyUpdate runs at a time.
	execUpdateMu sync.Mutex
}

// NewSelfUpdater returns nil if the app should not self-update.
func NewSelfUpdater(config *cfg.Config, eventsEmitter eventsEmitter) (*SelfUpdater, error) {
	if NoSelfUpdate == "true" {
		log.Printf("NoSelfUpdate=true; self-updates disabled")
		return nil, nil
	}

	if config == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.Version == "" {
		return nil, errors.New("cfg.Version is empty")
	}
	if eventsEmitter == nil {
		return nil, errors.New("eventsEmitter is nil")
	}

	if cfg.Version == "development" {
		log.Printf("cfg.Version=development; self-updates disabled")
		return nil, nil
	}

	httpClient := &http.Client{
		Timeout: httpTimeout,
	}

	u := SelfUpdater{
		version:       cfg.Version,
		config:        config,
		httpClient:    httpClient,
		eventsEmitter: eventsEmitter,
	}

	return &u, nil
}

func (su *SelfUpdater) RunScheduledUpdateChecks() {
	check := func() bool {
		policy := su.config.GetUpdatePolicy()
		if policy != cfg.UpdatePolicyAutomatic {
			return false
		}

		updated, err := su.execUpdate()
		if err != nil {
			log.Printf("failed to apply update: %v", err)
		}
		if updated {
			su.eventsEmitter.OnUpdateAvailable()
			return true
		}
		return false
	}

	if check() {
		return
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		if check() {
			return
		}
	}
}

// execUpdate checks and executes an update if available. Returns true if an update was applied.
func (su *SelfUpdater) execUpdate() (bool, error) {
	su.execUpdateMu.Lock()
	defer su.execUpdateMu.Unlock()

	rel, err := su.checkForUpdates()
	if err != nil {
		return false, fmt.Errorf("check for updates: %w", err)
	}
	if rel == nil {
		return false, nil
	}

	isNewer, err := su.isNewer(rel.Version)
	if err != nil {
		return false, fmt.Errorf("check if newer: %w", err)
	}
	if !isNewer {
		return false, nil
	}

	tmpFile, err := su.downloadAndVerifyFile(rel.AssetURL, rel.SHA256)
	if err != nil {
		return false, fmt.Errorf("download and verify file: %w", err)
	}
	defer os.Remove(tmpFile)

	switch runtime.GOOS {
	case "darwin":
		if err := su.applyUpdateForDarwin(tmpFile); err != nil {
			return false, fmt.Errorf("apply update: %w", err)
		}
	case "windows", "linux":
		if err := su.applyUpdateForWindowsOrLinux(tmpFile); err != nil {
			return false, fmt.Errorf("apply update: %w", err)
		}
	default:
		panic("unsupported platform")
	}

	log.Printf("update %s installed successfully", rel.Version)

	return true, nil
}

func (su *SelfUpdater) checkForUpdates() (*release, error) {
	log.Println("checking for updates")

	url := fmt.Sprintf("%s/%s/%s/%s/manifest.json", manifestsBaseURL, releaseTrack, runtime.GOOS, runtime.GOARCH)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "zen-desktop")

	resp, err := su.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed with status code %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &rel, nil
}

// isNewer compares the current version with the version passed as an argument and returns true if the argument is newer.
//
// It assumes that both versions are in the format "v<major>.<minor>.<patch>" and returns an error if they are not.
func (su *SelfUpdater) isNewer(version string) (bool, error) {
	var currentMajor, currentMinor, currentPatch, newMajor, newMinor, newPatch int
	if _, err := fmt.Sscanf(su.version, "v%d.%d.%d", &currentMajor, &currentMinor, &currentPatch); err != nil {
		return false, fmt.Errorf("parse current version (%s): %w", su.version, err)
	}
	if _, err := fmt.Sscanf(version, "v%d.%d.%d", &newMajor, &newMinor, &newPatch); err != nil {
		return false, fmt.Errorf("parse new version (%s): %w", version, err)
	}

	if newMajor > currentMajor {
		return true, nil
	}
	if newMajor == currentMajor && newMinor > currentMinor {
		return true, nil
	}
	if newMajor == currentMajor && newMinor == currentMinor && newPatch > currentPatch {
		return true, nil
	}

	return false, nil
}

func (su *SelfUpdater) downloadFile(url, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil) // #nosec G704 -- url comes from the update manifest, which is controlled by us
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Add("Accept", "application/octet-stream")

	resp, err := su.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download file failed with status code %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}

func verifyFileHash(filePath, expectedHash string) error {
	file, err := os.Open(filePath) // #nosec G703 -- filePath is generated by us and not influenced by user input
	if err != nil {
		return fmt.Errorf("open file for hashing: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("hash file: %w", err)
	}

	calculatedHash := hex.EncodeToString(hasher.Sum(nil))
	if calculatedHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, calculatedHash)
	}

	return nil
}

func (su *SelfUpdater) downloadAndVerifyFile(assetURL, expectedHash string) (string, error) {
	ext := filepath.Ext(assetURL)
	if strings.HasSuffix(assetURL, ".tar.gz") {
		ext = ".tar.gz"
	}

	if ext != ".tar.gz" && ext != ".zip" {
		return "", fmt.Errorf("unsupported archive format: %s", ext)
	}

	tmpFile, err := os.CreateTemp("", "downloaded-*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}

	if err := su.downloadFile(assetURL, tmpFile.Name()); err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}

	if err := verifyFileHash(tmpFile.Name(), expectedHash); err != nil {
		return "", fmt.Errorf("verify file hash: %w", err)
	}

	return tmpFile.Name(), nil
}

func (su *SelfUpdater) applyUpdateForDarwin(tmpFile string) error {
	tempDir, err := os.MkdirTemp("", "unarchive-*")
	if err != nil {
		return fmt.Errorf("create temp unarchive dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := unarchive(tmpFile, tempDir); err != nil {
		return fmt.Errorf("unarchive file: %w", err)
	}

	currentExecPath, err := getExecPath()
	if err != nil {
		return fmt.Errorf("get exec path: %w", err)
	}

	appBundlePath := findAppBundlePath(currentExecPath)
	if appBundlePath == "" {
		return fmt.Errorf("application is not running from an .app bundle: %s", currentExecPath)
	}

	newBundlePath, err := findAppBundleInDir(tempDir)
	if err != nil {
		return fmt.Errorf("find new app bundle: %w", err)
	}

	oldBundlePath := generateBackupName(appBundlePath)
	if err := os.Rename(appBundlePath, oldBundlePath); err != nil {
		return fmt.Errorf("rename current app bundle: %w", err)
	}

	rollback := false
	defer func() {
		if rollback {
			log.Printf("restoring old app bundle from: %s", oldBundlePath)

			if err := os.Rename(oldBundlePath, appBundlePath); err != nil {
				log.Printf("failed to restore old app bundle: %v", err)
			}
		} else {
			log.Printf("removing old app bundle backup: %s", oldBundlePath)

			if err := os.RemoveAll(oldBundlePath); err != nil {
				log.Printf("failed to remove old app bundle backup: %v", err)
			}
		}
	}()

	if err := os.Rename(newBundlePath, appBundlePath); err != nil {
		rollback = true
		return fmt.Errorf("rename new app bundle: %w", err)
	}

	return nil
}

func (su *SelfUpdater) applyUpdateForWindowsOrLinux(tmpFile string) error {
	tempDir, err := os.MkdirTemp("", "unarchive-*")
	if err != nil {
		return fmt.Errorf("create temp unarchive dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := unarchive(tmpFile, tempDir); err != nil {
		return fmt.Errorf("unzip file: %w", err)
	}

	currentExecPath, err := getExecPath()
	if err != nil {
		return fmt.Errorf("get exec path: %w", err)
	}

	oldExecPath := generateBackupName(currentExecPath)
	if err := os.Rename(currentExecPath, oldExecPath); err != nil {
		return fmt.Errorf("rename current executable to backup: %w", err)
	}

	rollback := false
	defer func() {
		if rollback {
			log.Printf("Restoring original executable from: %s", oldExecPath)
			if err := os.Rename(oldExecPath, currentExecPath); err != nil {
				log.Printf("Failed to restore original executable: %v", err)
			}
		} else {
			log.Printf("Removing backup executable: %s", oldExecPath)
			if err := os.Remove(oldExecPath); err != nil {
				log.Printf("Failed to remove backup executable: %v", err)

				log.Printf("Attempting to hide file: %s", oldExecPath)
				err = hideFile(oldExecPath)
				if err != nil {
					log.Printf("Failed to hide backup executable: %v", err)
				}
			}
		}
	}()

	if err := replaceExecutable(tempDir); err != nil {
		rollback = true
		return fmt.Errorf("replace executable: %w", err)
	}

	return nil
}

// hideFile moves the file at the given path to a temporary directory in case it cannot be removed.
func hideFile(path string) error {
	tmpDir := os.TempDir()
	newPath := filepath.Join(tmpDir, filepath.Base(path))

	if err := os.Rename(path, newPath); err != nil {
		return fmt.Errorf("move file to temp storage: %w", err)
	}

	log.Printf("moved file to temporary storage: %s", newPath)
	return nil
}

func findAppBundleInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && filepath.Ext(entry.Name()) == ".app" {
			return filepath.Join(dir, entry.Name()), nil
		}
	}
	return "", fmt.Errorf("no .app bundle found in directory: %s", dir)
}

func generateBackupName(originalName string) string {
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("%s.backup-%d", originalName, timestamp)
}
