package esxi

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePORTGROUPRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Println("[resourcePORTGROUPRead]")

	var vswitch string
	var vlan int
	var err error

	name := d.Id()

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
