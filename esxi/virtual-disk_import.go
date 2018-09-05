package esxi

import (
	"log"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVIRTUALDISKImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
  c := m.(*Config)
	log.Println("[resourceVIRTUALDISKImport]" )

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

  _, _, _, _, _, err := virtualDiskREAD(c, d.Id())
  if err != nil {
    d.SetId("")
    return results, fmt.Errorf("Failed to validate virtual_disk: %s\n", err)
	} else {
		d.SetId(d.Id())
  }

  return results, nil
}
