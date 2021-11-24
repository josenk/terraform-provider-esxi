package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePORTGROUPCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[resourcePORTGROUPCreate]")

	var stdout string
	var remote_cmd string
	var err error

	name := d.Get("name").(string)
	vswitch := d.Get("vswitch").(string)

	//  Create PORTGROUP
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard portgroup add -v \"%s\" -p \"%s\"",
		vswitch, name)

	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "create portgroup")
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to add portgroup: %s\n%s\n", stdout, err)
	}

	//  Set id
	d.SetId(name)

	return resourcePORTGROUPUpdate(d, m)
}
