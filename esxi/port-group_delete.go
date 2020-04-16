package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePortGroupDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourcePortGroupDelete]")

	var remote_cmd, cmd_result string
	var err error

	port_group_name := d.Get("port_group_name").(string)
	virtual_switch_name := getVirtualSwitchName(d.Get("virtual_switch_id").(string))

	remote_cmd = fmt.Sprintf(`vim-cmd hostsvc/net/portgroup_remove "%s" "%s"`, virtual_switch_name, port_group_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy port group")

	if err != nil {
		return fmt.Errorf("Unable to destroy port group: %s", err)
	}

	if cmd_result != "" {
		return fmt.Errorf("Unable to destroy port group: %s", cmd_result)
	}

	d.SetId("")
	return nil
}
