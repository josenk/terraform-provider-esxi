package esxi

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitchRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourceVirtualSwitchRead]")

	_, err := virtualSwitchRead(c, d.Id())
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("virtual_switch_name", d.Id())

	return nil
}
