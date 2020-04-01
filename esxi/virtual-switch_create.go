package esxi

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitchCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourceVirtualSwitchCreate]")

	virtual_switch_name := d.Get("virtual_switch_name").(string)

	if virtual_switch_name == "" {
		return errors.New("Virtual Switch Name must not be blank")

	}

	err := virtualSwitchCreate(c, virtual_switch_name)
	if err == nil {
		d.SetId(virtual_switch_name)
	} else {
		log.Println("[resourceVirtualSwitchCreate] Error: " + err.Error())
		d.SetId("")
		return fmt.Errorf("Failed to create virtual switch: %s\nError: %s", virtual_switch_name, err.Error())
	}

	return nil
}
