package esxi

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePortGroupRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourcePortGroupRead]")

	virtual_switch_name, port_group_name, err := portGroupRead(c, d.Get("port_group_name").(string), d.Get("virtual_switch_id").(string))
	if err != nil {
		d.SetId("")
		log.Println("[resourceVIRTUALDISKRead] Error: %s", err.Error())
		return err
	}

	d.Set("virtual_switch_name", virtual_switch_name)
	d.Set("port_group_name", port_group_name)

	return nil
}
