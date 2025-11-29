package sysproxy

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-multierror"
)

var (
	//go:embed exclusions/darwin.txt
	platformSpecificExcludedHosts []byte

	// networkServices remembers the services we modified, for unsetSystemProxy.
	networkServices []string
)

// setSystemProxy sets the system proxy PAC URL.
func setSystemProxy(pacURL string) error {
	svcs, err := discoverNetworkServices()
	if err != nil {
		return fmt.Errorf("discover network services: %v", err)
	}
	networkServices = svcs

	for _, svc := range networkServices {
		cmd := exec.Command("networksetup", "-setwebproxystate", svc, "off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("unset web proxy for network service %q: %v (%q)", svc, err, out)
		}

		cmd = exec.Command("networksetup", "-setsecurewebproxystate", svc, "off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("unset secure web proxy for network service %q: %v (%q)", svc, err, out)
		}

		cmd = exec.Command("networksetup", "-setautoproxyurl", svc, pacURL)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("set autoproxyurl to %q for network service %q: %v (%q)", pacURL, svc, err, out)
		}
	}

	return nil
}

func unsetSystemProxy() error {
	if len(networkServices) == 0 {
		return nil
	}

	var result error
	for _, svc := range networkServices {
		cmd := exec.Command("networksetup", "-setautoproxystate", svc, "off")
		if out, err := cmd.CombinedOutput(); err != nil {
			result = multierror.Append(result, fmt.Errorf("set autoproxystate to off for network service %q: %v (%q)", svc, err, out))
		}
	}

	networkServices = nil
	return result
}

// discoverNetworkServices returns a list of all network service names.
func discoverNetworkServices() ([]string, error) {
	cmd := exec.Command("networksetup", "-listallnetworkservices")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list network services: %v (%q)", err, out)
	}

	lines := bytes.Split(out, []byte{'\n'})
	if len(lines) < 2 {
		return nil, errors.New("no network services found")
	}

	// The first line contains "An asterisk (*) denotes that a network service is disabled."
	services := make([]string, 0, len(lines)-1)
	for _, raw := range lines[1:] {
		line := strings.TrimSpace(string(raw))
		if line == "" {
			continue
		}
		if line[0] == '*' {
			// Disabled service; remove the asterisk.
			line = strings.TrimSpace(line[1:])
		}

		services = append(services, line)
	}

	return services, nil
}
