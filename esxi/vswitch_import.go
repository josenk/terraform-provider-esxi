package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSWITCHImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)

	log.Println("[resourceVSWITCHImport]")

	var stdout string
	var err error

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

	// get vswitch (by name)
	remote_cmd := fmt.Sprintf("esxcli network vswitch standard list -v \"%s\"", d.Id())
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vswitch list")

	if err != nil {
		log.Printf("[resourceVSWITCHImport] Failed to import vswitch %s: %s\n", "vswitch list", err)
		return results, fmt.Errorf("Failed to import vswitch: %s\n%s\n", stdout, err)
	}

	d.SetId(d.Id())
	d.Set("name", d.Id())

	return results, nil
}
