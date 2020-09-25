package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSWITCHDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[resourceVSWITCHDelete]")

	var remote_cmd, stdout string
	var err error

	name := d.Id()

	remote_cmd = fmt.Sprintf("esxcli network vswitch standard remove -v \"%s\"", name)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "destroy vswitch")
	if err != nil {
		log.Printf("[resourceVSWITCHDelete] Failed destroy vswitch: %s\n", stdout)
		return fmt.Errorf("Failed to destroy vswitch: %s\n%s\n", stdout, err)
	}

	d.SetId("")
	return nil
}
