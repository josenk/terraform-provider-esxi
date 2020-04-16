package esxi

import (
	"fmt"
	"log"
	"regexp"
)

var virtual_switch_prefix = `key-vim.host.VirtualSwitch-`
var virtual_switch_id_regex = `"(key-vim\.host\.VirtualSwitch-.*)",`

func getVirtualSwitchName(virtual_switch_id string) string {
	return virtual_switch_id[27:]
}

func virtualSwitchCreate(c *Config, virtual_switch_name string) (string, error) {
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualSwitchCreate]")

	r := regexp.MustCompile(virtual_switch_id_regex)
	expected_virtual_switch_id := fmt.Sprintf("%s%s", virtual_switch_prefix, virtual_switch_name)
	var cmd_result, remote_cmd string
	var err error

	remote_cmd = fmt.Sprintf("vim-cmd hostsvc/net/vswitch_add \"%s\"", virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create virtual switch")

	if err != nil {
		return "", fmt.Errorf("Unable to create virtual switch: %s", err)
	}

	if cmd_result != "" {
		return "", fmt.Errorf("Unable to create virtual switch: %s", cmd_result)
	}

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/vswitch_info "%s"`, virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate virtual switch exists")

	if err != nil {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", err)
	}

	regex_output := r.FindStringSubmatch(cmd_result)

	if regex_output == nil || len(regex_output) < 2 {
		return "", fmt.Errorf("Unable to retrieve virtual switch id from 'vim-cmd hostsvc/net/vswitch_info' output")
	}

	virtual_switch_id := regex_output[1]

	if virtual_switch_id != expected_virtual_switch_id {
		return "", fmt.Errorf("Unable to validate virtual switch: Expected virtual switch id '%s' Actual virtual switch id '%s'", expected_virtual_switch_id, virtual_switch_id)
	}

	return virtual_switch_id, nil
}

func virtualSwitchRead(c *Config, virtual_switch_id string) (string, error) {

	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualSwitchRead] Begin")

	var remote_cmd, cmd_result string
	var err error

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/vswitch_info "%s"`, virtual_switch_id[27:])
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate virtual switch exists")

	if err != nil {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", err)
	}

	if cmd_result == "Not found." {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", cmd_result)
	}

	return virtual_switch_id[27:], nil
}

func virtualSwitchReadID(c *Config, virtual_switch_name string) (string, error) {

	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualSwitchRead] Begin")

	var remote_cmd, cmd_result string
	var err error

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/vswitch_info "%s"`, virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate virtual switch exists")

	if err != nil {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", err)
	}

	if cmd_result == "Not found." {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", cmd_result)
	}

	return fmt.Sprintf("%s%s", virtual_switch_prefix, virtual_switch_name), nil
}
