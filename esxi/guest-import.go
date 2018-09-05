package esxi

import (
	"log"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGUESTImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
  c := m.(*Config)
	log.Println("[resourceGUESTImport]" )

	var vmid string
	var err error

	results := make([]*schema.ResourceData, 1, 1)
  results[0] = d

	// get VMID (by name)
	vmid, err = guestValidateVMID(c, d.Id())
	if err != nil {
		return results, err
	}

	if vmid == d.Id() {
		d.SetId(vmid)
	} else {
		return results, fmt.Errorf("Failed to validate vmid: %s\n", err)
	}

  return results, nil
}
