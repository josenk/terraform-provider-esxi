package esxi

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func vswitchUpdate(c *Config, name string, ports int, mtu int, uplinks []string,
	link_discovery_mode string, promiscuous_mode bool, mac_changes bool, forged_transmits bool) error {
	esxiConnInfo := getConnectionInfo(c)

	log.Println("[vswitchUpdate]")

	var foundUplinks []string
	var remote_cmd, stdout string
	var err error

	//  Set mtu and cdp
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard set -m %d -c \"%s\" -v \"%s\"",
		mtu, link_discovery_mode, name)

	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "set vswitch mtu, link_discovery_mode")
	if err != nil {
		return fmt.Errorf("Failed to set vswitch mtu: %s\n%s\n", stdout, err)
	}

	//  Set security
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard policy security set -f %t -m %t -p %t -v \"%s\"",
		promiscuous_mode, mac_changes, forged_transmits, name)

	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "set vswitch security")
	if err != nil {
		return fmt.Errorf("Failed to set vswitch security: %s\n%s\n", stdout, err)
	}

	//  Update uplinks
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard list -v \"%s\"", name)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch list")

	if err != nil {
		log.Printf("[vswitchUpdate] Failed to run %s: %s\n", "vswitch list", err)
		return fmt.Errorf("Failed to list vswitch: %s\n%s\n", stdout, err)
	}

	re := regexp.MustCompile(`Uplinks: (.*)`)
	foundUplinksRaw := strings.Fields(re.FindStringSubmatch(stdout)[1])
	for i, s := range foundUplinksRaw {
		foundUplinks = append(foundUplinks, strings.Replace(s, ",", "", -1))
		log.Printf("[vswitchUpdate] found uplinks[%d]: /%s/\n", i, foundUplinks[i])
	}

	//  Add uplink if needed
	for i, _ := range uplinks {
		if inArrayOfStrings(foundUplinks, uplinks[i]) == false {
			log.Printf("[vswitchUpdate] add uplinks %d (%s)\n", i, uplinks[i])
			remote_cmd = fmt.Sprintf("esxcli network vswitch standard uplink add -u \"%s\" -v \"%s\"",
				uplinks[i], name)

			stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch add uplink")
			if strings.Contains(stdout, "Not a valid pnic") {
				return fmt.Errorf("Uplink not found: %s\n", uplinks[i])
			}
			if err != nil {
				return fmt.Errorf("Failed to add vswitch uplink: %s\n%s\n", stdout, err)
			}
		}
	}

	//  Remove uplink if needed
	for _, item := range foundUplinks {
		if inArrayOfStrings(uplinks, item) == false {
			log.Printf("[vswitchUpdate] delete uplink (%s)\n", item)
			remote_cmd = fmt.Sprintf("esxcli network vswitch standard uplink remove -u \"%s\" -v \"%s\"",
				item, name)

			stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch remove uplink")
			if err != nil {
				return fmt.Errorf("Failed to remove vswitch uplink: %s\n%s\n", stdout, err)
			}
		}
	}

	return nil
}

func vswitchRead(c *Config, name string) (int, int, []string, string, bool, bool, bool, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[vswitchRead]")

	var ports, mtu int
	var uplinks []string
	var link_discovery_mode string
	var promiscuous_mode, mac_changes, forged_transmits bool
	var remote_cmd, stdout string
	var err error

	remote_cmd = fmt.Sprintf("esxcli network vswitch standard list -v \"%s\"", name)
	stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch list")

	if stdout == "" {
		return 0, 0, uplinks, "", false, false, false, fmt.Errorf(stdout)
	}

	re, _ := regexp.Compile(`Configured Ports: ([0-9]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		ports, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		ports = 128
	}

	re, _ = regexp.Compile(`MTU: ([0-9]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		mtu, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		mtu = 1500
	}

	re, _ = regexp.Compile(`CDP Status: ([a-z]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		link_discovery_mode = re.FindStringSubmatch(stdout)[1]
	} else {
		link_discovery_mode = "listen"
	}

	re, _ = regexp.Compile(`Uplinks: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		foundUplinks := strings.Fields(re.FindStringSubmatch(stdout)[1])
		log.Printf("[vswitchRead] found foundUplinks: /%s/\n", foundUplinks)
		for i, s := range foundUplinks {
			uplinks = append(uplinks, strings.Replace(s, ",", "", -1))
			log.Printf("[vswitchRead] found uplinks[%d]: /%s/\n\n\n", i, uplinks[i])
		}
	} else {
		uplinks = uplinks[:0]
	}

	remote_cmd = fmt.Sprintf("esxcli network vswitch standard policy security get -v \"%s\"", name)
	stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch policy security get")

	if stdout == "" {
		log.Printf("[vswitchRead] Failed to run %s: %s\n", "vswitch policy security get", err)
		return 0, 0, uplinks, "", false, false, false, fmt.Errorf(stdout)
	}

	re, _ = regexp.Compile(`Allow Promiscuous: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		promiscuous_mode, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		promiscuous_mode = false
	}

	re, _ = regexp.Compile(`Allow MAC Address Change: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		mac_changes, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		mac_changes = false
	}

	re, _ = regexp.Compile(`Allow Forged Transmits: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		forged_transmits, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		forged_transmits = false
	}

	return ports, mtu, uplinks, link_discovery_mode, promiscuous_mode,
		mac_changes, forged_transmits, nil
}

//  Python is better... :-)
func inArrayOfStrings(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
