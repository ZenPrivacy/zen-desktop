package diskcache

import (
	"crypto/md5" // #nosec G501 -- MD5 is used to hash data, not for cryptographic purposes.
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const cacheTTL = 24 * time.Hour

type cacheEntry struct {
	Timestamp time.Time
	Content   []byte
}

type Cache struct {
	dir     string
	mu      sync.RWMutex
	entries map[string]cacheEntry // hash -> cache entry
}

func New() (*Cache, error) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	cache := &Cache{
		dir:     dir,
		entries: make(map[string]cacheEntry),
	}

	if err := cache.loadFromDisk(); err != nil {
		log.Printf("error loading cache from disk: %v", err)
	}

	return cache, nil
}

func getCacheDir() (string, error) {
	var appName string
	switch runtime.GOOS {
	case "darwin", "windows":
		appName = "Zen"
	case "linux":
		appName = "zen"
	default:
		panic("unsupported platform")
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName, "filters"), nil
}

func hashURL(url string) string {
	sum := md5.Sum([]byte(url)) // #nosec G401
	return hex.EncodeToString(sum[:])
}

func (c *Cache) loadFromDisk() error {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("read cache dir: %w", err)
	}

	for _, f := range files {
		name := f.Name()
		if !strings.HasSuffix(name, ".cache.txt") {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(f.Name(), ".cache.txt"), "-")
		if len(parts) != 2 {
			continue
		}

		hash := parts[0]
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		if time.Since(time.Unix(timestamp, 0)) > cacheTTL {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.dir, name))
		if err != nil {
			return fmt.Errorf("read cache file: %w", err)
		}

		c.entries[hash] = cacheEntry{
			Timestamp: time.Unix(timestamp, 0),
			Content:   data,
		}
	}

	return nil
}

func (c *Cache) Save(url string, content []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := hashURL(url)
	timestamp := time.Now()
	filename := fmt.Sprintf("%s-%d.cache.txt", hash, timestamp.Unix())

	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}

	fullPath := filepath.Join(c.dir, filename)
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return err
	}

	c.entries[hash] = cacheEntry{
		Timestamp: timestamp,
		Content:   content,
	}

	return nil
}

func (c *Cache) Load(url string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hash := hashURL(url)
	entry, ok := c.entries[hash]
	if !ok {
		return nil, false
	}

	if time.Since(entry.Timestamp) > cacheTTL {
		return nil, false
	}

	return entry.Content, true
}
