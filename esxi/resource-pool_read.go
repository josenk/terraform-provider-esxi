package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceRESOURCEPOOLRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Println("[resourceRESOURCEPOOLRead]")

	var cpu_shares, mem_shares string
	var cpu_min, cpu_max, mem_min, mem_max int
	var resource_pool_name, cpu_min_expandable, mem_min_expandable string
	var err error

	pool_id := d.Id()

	// Refresh
	resource_pool_name, cpu_min, cpu_min_expandable, cpu_max, cpu_shares, mem_min, mem_min_expandable, mem_max, mem_shares, err = resourcePoolRead(c, pool_id)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("resource_pool_name", resource_pool_name)
	d.Set("cpu_min", cpu_min)
	d.Set("cpu_min_expandable", cpu_min_expandable)
	d.Set("cpu_max", cpu_max)
	d.Set("cpu_shares", cpu_shares)
	d.Set("mem_min", mem_min)
	d.Set("mem_min_expandable", mem_min_expandable)
	d.Set("mem_max", mem_max)
	d.Set("mem_shares", mem_shares)

	return nil
}
