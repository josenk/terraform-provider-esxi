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
		// check if message on error is because a portgroup is already existing on host
		var msg_already_exists string
		msg_already_exists = fmt.Sprintf("A portgroup with the name %s already exists", name)
		if stdout != msg_already_exists {
			d.SetId("")
			return fmt.Errorf("Failed to add portgroup: %s\n%s\n", stdout, err)
		} else {
			log.Printf("[resourcePORTGROUPCreate] %s - continuing\n", stdout)
		}
	}

	//  Set id
	d.SetId(name)

	return resourcePORTGROUPUpdate(d, m)
}
