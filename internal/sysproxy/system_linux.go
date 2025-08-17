package sysproxy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	platformSpecificExcludedHosts []byte
)

func detectDesktopEnvironment() string {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")

	if strings.ToLower(desktop) == "kde" {
		return "kde"
	}

	if strings.Contains(strings.ToLower(desktop), "gnome") {
		return "gnome"
	}

	return ""
}

func setSystemProxy(pacURL string) error {
	desktop := detectDesktopEnvironment()

	// TODO: add support for other desktop environments
	switch desktop {
	case "kde":
		if err := setKDEProxy(pacURL); err != nil {
			return err
		}

		// Set gsettings on KDE as firefox based browsers ignores KDE proxy
		if binaryExists("gsettings") {
			if err := setGnomeProxy(pacURL); err != nil {
				return err
			}
		}
		return nil

	case "gnome":
		return setGnomeProxy(pacURL)

	default:
		return ErrUnsupportedDesktopEnvironment
	}
}

func setKDEProxy(pacURL string) error {
	var kwriteconfig string

	if binaryExists("kwriteconfig6") {
		kwriteconfig = "kwriteconfig6"
	}

	if binaryExists("kwriteconfig5") {
		kwriteconfig = "kwriteconfig5"
	}

	commands := [][]string{
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ProxyType", "2"},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "Proxy Config Script", pacURL},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ReversedException", "false"},
	}

	for _, command := range commands {
		cmd := exec.Command(command[0], command[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run KDE proxy command %q: %v (%q)", strings.Join(command, " "), err, out)
		}
	}

	return nil
}

func setGnomeProxy(pacURL string) error {

	commands := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "autoconfig-url", pacURL},
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "auto"},
	}

	for _, command := range commands {
		cmd := exec.Command(command[0], command[1:]...) // #nosec G204
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run GNOME proxy command %q: %v (%q)", strings.Join(command, " "), err, out)
		}
	}
	return nil
}

func unsetSystemProxy() error {

	desktop := detectDesktopEnvironment()
	switch desktop {
	case "kde":
		if err := unsetKDEProxy(); err != nil {
			return err
		}

		if binaryExists("gsettings") {
			if err := unsetGnomeProxy(); err != nil {
				return err
			}
		}
		return nil

	case "gnome":
		return unsetGnomeProxy()

	default:
		return ErrUnsupportedDesktopEnvironment
	}
}

func unsetKDEProxy() error {

	var kwriteconfig string

	if binaryExists("kwriteconfig6") {
		kwriteconfig = "kwriteconfig6"
	}

	if binaryExists("kwriteconfig5") {
		kwriteconfig = "kwriteconfig5"
	}

	commands := [][]string{
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ProxyType", "0"},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "Proxy Config Script", ""},
	}

	for _, command := range commands {
		cmd := exec.Command(command[0], command[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("unset KDE proxy: %v (%q)", err, out)
		}
	}

	return nil
}

func unsetGnomeProxy() error {

	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unset GNOME proxy: %v (%q)", err, out)
	}

	return nil
}

func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
