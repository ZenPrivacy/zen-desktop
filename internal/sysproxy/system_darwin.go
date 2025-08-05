package sysproxy

import (
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/getlantern/elevate"
)

var (
	reInterfaceName = regexp.MustCompile(`^[\w\d]+$`)
	networkService  string
	//go:embed exclusions/darwin.txt
	platformSpecificExcludedHosts []byte
)

type command struct {
	name string
	args []string
}

func (c command) String() string {
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

// setSystemProxy sets the system proxy to the proxy address.
func setSystemProxy(pacURL string) error {
	cmd := exec.Command("sh", "-c", "scutil --nwi | grep 'Network interfaces' | cut -d ' ' -f 3")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("get default interface: %v\n%s", err, out)
	}
	interfaceName := strings.TrimSpace(string(out))
	if len(interfaceName) == 0 {
		return errors.New("no default interface found")
	}
	if !reInterfaceName.MatchString(interfaceName) {
		// I am pretty sure that interface names can only contain alphanumeric characters,
		// but just to be sure not to introduce a shell injection vulnerability, let's check it.
		return fmt.Errorf("invalid interface name: %s", interfaceName)
	}

	cmd = exec.Command("sh", "-c", fmt.Sprintf("networksetup -listnetworkserviceorder | grep %s -B 1 | head -n 1 | cut -d ' ' -f 2-", interfaceName)) // #nosec G204 -- Interface name is validated above
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("get network service: %v\n%s", err, out)
	}
	networkService = strings.TrimSpace(string(out))
	if len(networkService) == 0 {
		return errors.New("no network service found")
	}

	cmds := []command{
		{
			name: "networksetup",
			args: []string{"-setwebproxystate", networkService, "off"},
		},
		{
			name: "networksetup",
			args: []string{"-setsecurewebproxystate", networkService, "off"},
		},
		{
			name: "networksetup",
			args: []string{"-setautoproxyurl", networkService, pacURL},
		},
	}

	var retryCommands []string
	for _, c := range cmds {
		cmd := exec.Command(c.name, c.args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			retryCommands = append(retryCommands, c.String())
		}
	}

	if retryCommands != nil {
		cmd = elevate.WithPrompt("System changes required to activate proxy").Command("sh", "-c", strings.Join(retryCommands, " "))
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("network setup with root privileges: %w (%q)", err, out)
		}
	}

	// There's no need to set autoproxystate to on, as setting the URL already does that.

	return nil
}

func unsetSystemProxy() error {
	if networkService == "" {
		return errors.New("trying to unset system proxy without setting it first")
	}

	cmd := command{
		name: "networksetup",
		args: []string{"-setautoproxystate", networkService, "off"},
	}

	var retry bool
	c := exec.Command(cmd.name, cmd.args...)
	if _, err := c.CombinedOutput(); err != nil {
		retry = true
	}

	if retry {
		cmd := elevate.WithPrompt("System changes required to deactivate proxy").Command(cmd.name, cmd.args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("proxy deactivation with root privileges: %w (%q)", err, out)
		}
	}

	networkService = ""

	return nil
}
