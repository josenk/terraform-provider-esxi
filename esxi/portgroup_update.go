package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePORTGROUPUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[resourcePORTGROUPUpdate]")

	var stdout string
	var remote_cmd string
	var err error

	name := d.Get("name").(string)
	vswitch := d.Get("vswitch").(string)
	vlan := d.Get("vlan").(int)

	//  set vlan id
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard portgroup set -v \"%d\" -p \"%s\"",
		vlan, name)

	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup set vlan")
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to set portgroup: %s\n%s\n", stdout, err)
	}

	// Refresh
	vswitch, vlan, err = portgroupRead(c, name)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("vswitch", vswitch)
	d.Set("vlan", vlan)

	return nil
}
