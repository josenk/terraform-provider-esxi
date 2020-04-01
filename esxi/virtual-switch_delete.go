package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitchDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceVirtualSwitchDelete]")

	var remote_cmd, cmd_result string
	var err error

	virtual_switch_name := d.Id()

	remote_cmd = fmt.Sprintf("esxcfg-vswitch -d \"%s\"", virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy virtual switch")

	if err != nil {
		return fmt.Errorf("Unable to destroy virtual switch: %w", err)
	}

	if cmd_result != "" {
		return fmt.Errorf("Unable to destroy virtual switch: %s", cmd_result)
	}

	d.SetId("")
	return nil
}
