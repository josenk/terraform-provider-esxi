package esxi

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePortGroupCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourcePortGroupCreate]")

	port_group_name := d.Get("port_group_name").(string)

	if port_group_name == "" {
		return errors.New("Port Group Name must not be blank")
	}

	virtual_switch_id := d.Get("virtual_switch_id").(string)

	if virtual_switch_id == "" {
		return errors.New("Virtual Switch ID must not be blank")
	}

	_, err := virtualSwitchRead(c, virtual_switch_id)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to validate virtual_switch: %w", err)
	}

	port_group_id, err := portGroupCreate(c, virtual_switch_id, port_group_name)
	if err == nil {
		d.SetId(port_group_id)
	} else {
		log.Println("[resourcePortGroupCreate] Error: " + err.Error())
		d.SetId("")
		return fmt.Errorf("Failed to create port group: %s Error: %s", port_group_name, err.Error())
	}

	return nil
}
