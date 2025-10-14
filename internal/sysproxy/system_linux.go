package sysproxy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var platformSpecificExcludedHosts []byte

func detectDesktopEnvironment() string {
	xdg := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(xdg, "kde") || strings.Contains(xdg, "plasma") {
		return "kde"
	}

	if strings.Contains(xdg, "gnome") {
		return "gnome"
	}

	// Fallback DE checks
	ds := strings.ToLower(os.Getenv("DESKTOP_SESSION"))
	if strings.Contains(ds, "kde") || strings.Contains(ds, "plasma") || strings.ToLower(os.Getenv("KDE_FULL_SESSION")) == "true" {
		return "kde"
	}

	if strings.Contains(ds, "gnome") {
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
				log.Printf("warning: failed to apply GNOME proxy on KDE: %v", err)
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
	} else if binaryExists("kwriteconfig5") {
		kwriteconfig = "kwriteconfig5"
	}
	if kwriteconfig == "" {
		return fmt.Errorf("kwriteconfig not found in PATH, cannot configure KDE proxy")
	}

	commands := [][]string{
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ProxyType", "2"},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "Proxy Config Script", pacURL},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ReversedException", "false"},
	}

	for _, command := range commands {
		out, err := runCmdWithTimeout(command[0], command[1:]...)
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
		out, err := runCmdWithTimeout(command[0], command[1:]...)
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
				log.Printf("warning: failed to unset GNOME proxy on KDE: %v", err)
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
	} else if binaryExists("kwriteconfig5") {
		kwriteconfig = "kwriteconfig5"
	}

	if kwriteconfig == "" {
		return fmt.Errorf("kwriteconfig not found in PATH, cannot unset KDE proxy")
	}

	commands := [][]string{
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ProxyType", "0"},
		{kwriteconfig, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "Proxy Config Script", ""},
	}

	for _, command := range commands {
		out, err := runCmdWithTimeout(command[0], command[1:]...)
		if err != nil {
			return fmt.Errorf("unset KDE proxy: %v (%q)", err, out)
		}
	}

	return nil
}

func unsetGnomeProxy() error {
	commands := [][]string{
		{"gsettings", "reset", "org.gnome.system.proxy", "autoconfig-url"},
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "none"},
	}
	for _, command := range commands {
		out, err := runCmdWithTimeout(command[0], command[1:]...)
		if err != nil {
			return fmt.Errorf("unset GNOME proxy: %v (%q)", err, out)
		}
	}

	return nil
}

func runCmdWithTimeout(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204
	return cmd.CombinedOutput()
}

func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
