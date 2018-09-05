package esxi

import (
	"log"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)


func resourceRESOURCEPOOLImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
  c := m.(*Config)

	log.Println("[resourceRESOURCEPOOLImport]" )

	var err error

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

	// get VMID (by name)
	_, err = getPoolNAME(c, d.Id())
	if err != nil {
		return results, fmt.Errorf("Failed to validate resource_pool: %s\n", err)
	} else {
		d.SetId(d.Id())
	}

  return results, nil
}
