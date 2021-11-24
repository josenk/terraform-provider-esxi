package esxi

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSWITCHRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Println("[resourceVSWITCHRead]")

	var ports, mtu int
	var uplinks []string
	var link_discovery_mode string
	var promiscuous_mode, mac_changes, forged_transmits bool
	var err error

	name := d.Id()

	// Refresh
	ports, mtu, uplinks, link_discovery_mode, promiscuous_mode, mac_changes, forged_transmits, err = vswitchRead(c, name)
	if err != nil {
		d.SetId("")
		return nil
	}

	// Change uplinks (list) to map
	log.Printf("[resourceVSWITCHRead] uplinks: %q\n", uplinks)
	uplink := make([]map[string]interface{}, 0, 1)

	if len(uplinks) == 0 {
		uplink = nil
	} else {
		for i, _ := range uplinks {
			out := make(map[string]interface{})
			out["name"] = uplinks[i]
			uplink = append(uplink, out)
		}
	}
	d.Set("uplink", uplink)

	d.Set("ports", ports)
	d.Set("mtu", mtu)
	d.Set("link_discovery_mode", link_discovery_mode)
	d.Set("promiscuous_mode", promiscuous_mode)
	d.Set("mac_changes", mac_changes)
	d.Set("forged_transmits", forged_transmits)

	return nil
}
