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
	vlan := d.Get("vlan").(int)

	//  set vlan id
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard portgroup set -v \"%d\" -p \"%s\"",
		vlan, name)

	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup set vlan")
	if err != nil {
		d.SetId("") // TODO do we really want to do this? maybe only if the portgroup
		return fmt.Errorf("Failed to set portgroup: %s\n%s\n", stdout, err)
	}

	// set the security policy.
	promiscuous_mode := d.Get("promiscuous_mode").(bool)
	forged_transmits := d.Get("forged_transmits").(bool)
	mac_changes := d.Get("mac_changes").(bool)
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard portgroup policy security set -p \"%s\" --allow-promiscuous=%t --allow-forged-transmits=%t --allow-mac-change=%t", name, promiscuous_mode, forged_transmits, mac_changes)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup set vlan")
	if err != nil {
		return fmt.Errorf("Failed to set the portgroup security policy: %s\n%s\n", stdout, err)
	}

	// Refresh
	return resourcePORTGROUPRead(d, m)
}
