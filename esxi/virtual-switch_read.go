package esxi

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitchRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourceVirtualSwitchRead]")

	virtual_switch_name, err := virtualSwitchRead(c, d.Id())
	if err != nil {
		d.SetId("")
		log.Println("[resourceVirtualSwitchRead] Error: %s", err.Error())
		return err
	}

	d.Set("virtual_switch_name", virtual_switch_name)

	return nil
}
