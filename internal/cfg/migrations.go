package cfg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ZenPrivacy/zen-desktop/internal/autostart"
	"github.com/blang/semver"
)

// migrations is a map of version to migration function.
// Warning: RunMigration() runs the migrations in arbitrary order.
var migrations = map[string]func(c *Config) error{
	"v0.9.0": func(c *Config) error {
		c.Lock()
		defer c.Unlock()

		c.UpdatePolicy = UpdatePolicyPrompt
		if err := c.Save(); err != nil {
			return fmt.Errorf("save config: %v", err)
		}

		if runtime.GOOS != "darwin" {
			autostart := autostart.Manager{}
			if enabled, err := autostart.IsEnabled(); err != nil {
				return fmt.Errorf("check enabled: %w", err)
			} else if enabled {
				// Re-enable to change autostart command
				if err := autostart.Disable(); err != nil {
					return fmt.Errorf("disable autostart: %w", err)
				}
				if err := autostart.Enable(); err != nil {
					return fmt.Errorf("enable autostart: %w", err)
				}
			}
		}

		return nil
	},
	"v0.11.0": func(c *Config) error {
		oldFilterLists := c.GetFilterLists()
		var customFilterLists []CustomFilterList

		for _, list := range oldFilterLists {
			if list.Type == FilterListTypeCustom {
				customFilterLists = append(customFilterLists, CustomFilterList{
					Name:    list.Name,
					URL:     list.URL,
					Enabled: list.Enabled,
					Trusted: list.Trusted,
				})
			}
		}
		if len(customFilterLists) == 0 {
			return nil
		}

		if err := c.AddCustomFilterLists(customFilterLists); err != nil {
			return fmt.Errorf("add custom filter lists: %v", err)
		}
		return nil
	},
}

// RunMigrations runs the version-to-version migrations.
func (c *Config) RunMigrations() {
	if Version == "development" {
		log.Println("skipping migrations in development mode")
		return
	}

	var lastMigration string
	lastMigrationFile := filepath.Join(ConfigDir, "last_migration")
	if c.firstLaunch {
		lastMigration = Version
	} else {
		if _, err := os.Stat(lastMigrationFile); !os.IsNotExist(err) {
			lastMigrationData, err := os.ReadFile(lastMigrationFile)
			if err != nil {
				log.Fatalf("failed to read last migration file: %v", err)
			}
			lastMigration = string(lastMigrationData)
		} else {
			// Should trigger when updating from pre v0.3.0
			lastMigration = "v0.0.0"
		}
	}

	lastMigrationV, err := semver.ParseTolerant(lastMigration)
	if err != nil {
		log.Printf("error parsing last migration(%s): %v\n", lastMigration, err)
		return
	}

	for version, migration := range migrations {
		versionV, err := semver.ParseTolerant(version)
		if err != nil {
			log.Printf("error parsing migration version(%s): %v\n", version, err)
			continue
		}

		if lastMigrationV.LT(versionV) {
			if err := migration(c); err != nil {
				log.Printf("error running migration(%s): %v\n", version, err)
			} else {
				log.Printf("ran migration %s\n", version)
			}
		}
	}

	if err := os.WriteFile(lastMigrationFile, []byte(Version), 0644); err != nil {
		log.Printf("error writing last migration file: %v", err)
	}
}
