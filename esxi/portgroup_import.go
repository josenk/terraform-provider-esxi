package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePORTGROUPImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)

	log.Println("[resourcePORTGROUPImport]")

	var stdout string
	var err error

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

	// get porgroup (by name)
	remote_cmd := fmt.Sprintf("esxcli network vswitch standard portgroup list |grep -m 1 \"^%s\"", d.Id())
	_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup list")

	if err != nil {
		log.Printf("[resourceVSWITCHImport] Failed to import portgroup %s: %s\n", "vswitch list", err)
		return results, fmt.Errorf("Failed to import portgroup: %s\n%s\n", stdout, err)
	}

	d.SetId(d.Id())
	d.Set("name", d.Id())

	return results, nil
}
