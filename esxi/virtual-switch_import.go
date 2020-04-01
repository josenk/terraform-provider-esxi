package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitchImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*Config)
	log.Println("[resourceVirtualSwitchImport]")

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

	_, err := virtualSwitchRead(c, d.Id())
	if err != nil {
		d.SetId("")
		return results, fmt.Errorf("Failed to validate virtual_switch: %s", err)
	}

	d.SetId(d.Id())

	return results, nil
}
