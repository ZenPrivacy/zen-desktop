package cfg

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	// ConfigDir is the path to the directory storing the application configuration.
	ConfigDir string
	// DataDir is the path to the directory storing the application data.
	DataDir string
	// Version is the current version of the application. Set at compile time for production builds using ldflags (see tasks in the /tasks/build directory).
	Version = "development"
)

//go:embed default-config.json
var defaultConfig embed.FS

type FilterListType string

const (
	FilterListTypeCustom FilterListType = "custom"
)

type UpdatePolicyType string

const (
	UpdatePolicyAutomatic UpdatePolicyType = "automatic"
	UpdatePolicyPrompt    UpdatePolicyType = "prompt"
	UpdatePolicyDisabled  UpdatePolicyType = "disabled"
)

type BuiltInPref struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// Config stores and manages the configuration for the application.
// Although all fields are public, this is only for use by the JSON marshaller.
// All access to the Config should be done through the exported methods.
type Config struct {
	sync.RWMutex

	Filter struct {
		// FilterLists was a unified list of filter lists, including built-in and custom lists.
		//
		// Deprecated: FilterLists exists to facilitate a migration for users upgrading from pre-v0.11.0. Use CustomLists and BuiltInConfig instead.
		FilterLists []FilterList `json:"filterLists"`
		// CustomLists are user-defined lists.
		CustomLists []FilterList `json:"customLists"`
		// BuiltInPrefs stores user preferences for built-in filter lists. Enabled has a priority over builtInFilterList.DefaultEnabled.
		BuiltInPrefs []BuiltInPref `json:"builtInConfig"`
		MyRules      []string      `json:"myRules"`
	} `json:"filter"`
	Certmanager struct {
		CAInstalled bool `json:"caInstalled"`
	} `json:"certmanager"`
	Proxy struct {
		Port         int      `json:"port"`
		IgnoredHosts []string `json:"ignoredHosts"`
		PACPort      int      `json:"pacPort"`
	} `json:"proxy"`
	UpdatePolicy UpdatePolicyType `json:"updatePolicy"`

	Locale string `json:"locale"`

	// firstLaunch is true if the application is being run for the first time.
	firstLaunch bool
}

type FilterList struct {
	Name    string         `json:"name"`
	Type    FilterListType `json:"type"`
	URL     string         `json:"url"`
	Enabled bool           `json:"enabled"`
	Trusted bool           `json:"trusted"`
}

func (f *FilterList) UnmarshalJSON(data []byte) error {
	type TempFilterList FilterList
	var temp TempFilterList

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp.Name == "" {
		return errors.New("name is required")
	}

	if temp.URL == "" {
		return errors.New("URL is required")
	}

	if temp.Type == "" {
		return errors.New("type is required")
	}

	*f = FilterList(temp)
	return nil
}

func init() {
	var err error
	ConfigDir, err = getConfigDir()
	if err != nil {
		log.Fatalf("failed to get config dir: %v", err)
	}
	stat, err := os.Stat(ConfigDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(ConfigDir, 0755); err != nil {
			log.Fatalf("failed to create config dir: %v", err)
		}
		stat, err = os.Stat(ConfigDir)
	}
	if err != nil {
		log.Fatalf("failed to stat config dir: %v", err)
	}
	if !stat.IsDir() {
		log.Fatalf("config dir is not a directory: %s", ConfigDir)
	}

	DataDir, err = getDataDir()
	if err != nil {
		log.Fatalf("failed to get data dir: %v", err)
	}
	stat, err = os.Stat(DataDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(DataDir, 0755); err != nil {
			log.Fatalf("failed to create data dir: %v", err)
		}
		stat, err = os.Stat(DataDir)
	}
	if err != nil {
		log.Fatalf("failed to stat data dir: %v", err)
	}
	if !stat.IsDir() {
		log.Fatalf("data dir is not a directory: %s", DataDir)
	}
}

func NewConfig() (*Config, error) {
	c := &Config{}

	configFile := filepath.Join(ConfigDir, "config.json")
	var configData []byte
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		configData, err = os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	} else {
		configData, err = defaultConfig.ReadFile("default-config.json")
		if err != nil {
			return nil, fmt.Errorf("failed to read default config file: %v", err)
		}
		if err := os.WriteFile(configFile, configData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write config file: %v", err)
		}
		c.firstLaunch = true
	}

	if err := json.Unmarshal(configData, c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return c, nil
}

// Save saves the config to disk.
// It is not thread-safe, and should only be called if the caller has
// a lock on the config.
func (c *Config) Save() error {
	configData, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configFile := filepath.Join(ConfigDir, "config.json")
	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func (c *Config) isCustomURL(url string) bool {
	for _, list := range c.Filter.CustomLists {
		if list.URL == url {
			return true
		}
	}
	return false
}

// GetFilterLists returns the list of filter lists.
func (c *Config) GetFilterLists() []FilterList {
	c.RLock()
	defer c.RUnlock()

	res := make([]FilterList, 0, len(c.Filter.CustomLists)+len(builtInFilterLists))
	for _, builtIn := range builtInFilterLists {
		enabled := builtIn.DefaultEnabled
		for _, pref := range c.Filter.BuiltInPrefs {
			if pref.URL != builtIn.URL {
				continue
			}
			enabled = pref.Enabled
			break
		}
		res = append(res, FilterList{
			Name:    builtIn.Name,
			Type:    builtIn.Type,
			URL:     builtIn.URL,
			Enabled: enabled,
			Trusted: builtIn.Trusted,
		})
	}

	res = append(res, c.Filter.CustomLists...)
	return res
}

// AddFilterList adds a filter list to the custom filter lists.
func (c *Config) AddFilterList(list FilterList) error {
	c.Lock()
	defer c.Unlock()

	if isBuiltInURL(list.URL) {
		return errors.New("this list already exists in the built-in configuration")
	}
	if c.isCustomURL(list.URL) {
		return errors.New("this list already exists in the custom configuration")
	}

	c.Filter.CustomLists = append(c.Filter.CustomLists, list)
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
		return fmt.Errorf("save: %v", err)
	}
	return nil
}

// AddFilterLists adds multiple filter lists to the configuration.
func (c *Config) AddFilterLists(lists []FilterList) error {
	c.Lock()
	defer c.Unlock()

	newLists := make([]FilterList, 0, len(lists))
	for _, list := range lists {
		if isBuiltInURL(list.URL) {
			log.Printf("adding filter lists: list %s already exists in the built-in configuration", list.URL)
			continue
		}
		if c.isCustomURL(list.URL) {
			log.Printf("adding filter lists: list %s already exists in the custom configuration", list.URL)
			continue
		}

		newLists = append(newLists, list)
	}

	c.Filter.CustomLists = append(c.Filter.CustomLists, newLists...)
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
		return fmt.Errorf("save: %v", err)
	}
	return nil
}

// RemoveFilterList removes a filter list from the list of enabled filter lists.
func (c *Config) RemoveFilterList(url string) error {
	c.Lock()
	defer c.Unlock()

	var found bool
	for i, list := range c.Filter.CustomLists {
		if list.URL == url {
			c.Filter.FilterLists = append(c.Filter.FilterLists[:i], c.Filter.FilterLists[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("filter list not found")
	}
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
		return fmt.Errorf("save: %v", err)
	}
	return nil
}

func (c *Config) ToggleFilterList(url string, enabled bool) error {
	c.Lock()
	defer c.Unlock()

	if c.toggleBuiltIn(url, enabled) || c.toggleCustom(url, enabled) {
		if err := c.Save(); err != nil {
			log.Printf("failed to save config: %v", err)
			return fmt.Errorf("save: %v", err)
		}
		return nil
	}
	return errors.New("filter list not found")
}

func (c *Config) toggleBuiltIn(url string, enabled bool) bool {
	if !isBuiltInURL(url) {
		return false
	}
	for i, pref := range c.Filter.BuiltInPrefs {
		if pref.URL == url {
			c.Filter.BuiltInPrefs[i].Enabled = enabled
			return true
		}
	}
	c.Filter.BuiltInPrefs = append(c.Filter.BuiltInPrefs, BuiltInPref{URL: url, Enabled: enabled})
	return true
}

func (c *Config) toggleCustom(url string, enabled bool) bool {
	for i, list := range c.Filter.CustomLists {
		if list.URL == url {
			c.Filter.CustomLists[i].Enabled = enabled
			return true
		}
	}
	return false
}

func (c *Config) GetCustomFilterLists() []FilterList {
	c.RLock()
	defer c.RUnlock()

	return c.Filter.CustomLists
}

func (c *Config) GetMyRules() []string {
	c.RLock()
	defer c.RUnlock()

	return c.Filter.MyRules
}

func (c *Config) SetMyRules(rules []string) error {
	c.Lock()
	defer c.Unlock()

	c.Filter.MyRules = rules
	if err := c.Save(); err != nil {
		err = fmt.Errorf("failed to save config: %v", err)
		log.Println(err)
		return err
	}
	return nil
}

// GetPort returns the port the proxy is set to listen on.
func (c *Config) GetPort() int {
	c.RLock()
	defer c.RUnlock()

	return c.Proxy.Port
}

// SetPort sets the port the proxy is set to listen on.
func (c *Config) SetPort(port int) string {
	c.Lock()
	defer c.Unlock()

	c.Proxy.Port = port
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
		return err.Error()
	}
	return ""
}

// GetIgnoredHosts returns the list of ignored hosts.
func (c *Config) GetIgnoredHosts() []string {
	c.RLock()
	defer c.RUnlock()

	return c.Proxy.IgnoredHosts
}

// SetIgnoredHosts sets the list of ignored hosts.
func (c *Config) SetIgnoredHosts(hosts []string) error {
	c.Lock()
	defer c.Unlock()

	c.Proxy.IgnoredHosts = hosts
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
		return err
	}
	return nil
}

// GetCAInstalled returns whether the CA is installed.
func (c *Config) GetCAInstalled() bool {
	c.RLock()
	defer c.RUnlock()

	return c.Certmanager.CAInstalled
}

// SetCAInstalled sets whether the CA is installed.
func (c *Config) SetCAInstalled(caInstalled bool) {
	c.Lock()
	defer c.Unlock()

	c.Certmanager.CAInstalled = caInstalled
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
	}
}

func (c *Config) GetPACPort() int {
	c.RLock()
	defer c.RUnlock()

	return c.Proxy.PACPort
}

func (c *Config) GetVersion() string {
	return Version
}

func (c *Config) GetUpdatePolicy() UpdatePolicyType {
	c.RLock()
	defer c.RUnlock()

	return c.UpdatePolicy
}

func (c *Config) SetUpdatePolicy(p UpdatePolicyType) {
	c.Lock()
	defer c.Unlock()

	c.UpdatePolicy = p
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
	}
}

func (c *Config) GetLocale() string {
	c.RLock()
	defer c.RUnlock()

	return c.Locale
}

func (c *Config) SetLocale(l string) {
	c.Lock()
	defer c.Unlock()

	c.Locale = l
	if err := c.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
	}
}
