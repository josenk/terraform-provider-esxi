package esxi

import (
	"fmt"
	"log"
	"regexp"
)

var port_group_prefix = `key-vim.host.PortGroup-`
var port_group_id_regex = `(key-vim\.host\.PortGroup-.*)>`

func getPortGroupName(port_group_id string) string {
	return port_group_id[23:]
}

func portGroupCreate(c *Config, virtual_switch_id string, port_group_name string) (string, error) {
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[portGroupCreate]")

	r := regexp.MustCompile(port_group_id_regex)
	expected_port_group_id := fmt.Sprintf("%s%s", port_group_prefix, port_group_name)
	var cmd_result, remote_cmd string
	var err error

	remote_cmd = fmt.Sprintf(`esxcfg-vswitch -C "%s"`, port_group_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate port group does not exist")

	if err != nil {
		return "", fmt.Errorf("Unable to validate port group: %s", err)
	}

	if cmd_result == "1" {
		return "", fmt.Errorf("Port group already exists")
	}

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/portgroup_add "%s" "%s"`, getVirtualSwitchName(virtual_switch_id), port_group_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create port group")

	if err != nil {
		return "", fmt.Errorf("Unable to create port group: %s", err)
	}

	if cmd_result != "" {
		return "", fmt.Errorf("Unable to create port group: %s", cmd_result)
	}

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/vswitch_info "%s"`, getVirtualSwitchName(virtual_switch_id))
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate port group exists")

	if err != nil {
		return "", fmt.Errorf("Unable to validate port group: %s", err)
	}

	regex_output := r.FindStringSubmatch(cmd_result)

	if regex_output == nil || len(regex_output) < 2 {
		return "", fmt.Errorf("Unable to retrieve port group id from 'vim-cmd hostsvc/net/vswitch_info' output")
	}

	port_group_id := regex_output[1]

	if port_group_id != expected_port_group_id {
		return "", fmt.Errorf("Unable to validate port group: Expected port group id '%s' Actual port group id '%s'", expected_port_group_id, port_group_id)
	}

	return port_group_id, nil
}

func portGroupRead(c *Config, virtual_switch_id string) (string, string, error) {

	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[portGroupRead] Begin")

	r := regexp.MustCompile(port_group_id_regex)
	var remote_cmd, cmd_result string
	var err error

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/vswitch_info "%s"`, getVirtualSwitchName(virtual_switch_id))
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate port group exists")

	if err != nil {
		return "", "", fmt.Errorf("Unable to validate port group: %s", err)
	}

	regex_output := r.FindStringSubmatch(cmd_result)

	if regex_output == nil || len(regex_output) < 2 {
		return "", "", fmt.Errorf("Unable to retrieve port group id from 'vim-cmd hostsvc/net/vswitch_info' output")
	}

	port_group_id := regex_output[1]

	return getVirtualSwitchName(virtual_switch_id), getPortGroupName(port_group_id), nil
}
