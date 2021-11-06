package esxi

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSWITCHCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[resourceVSWITCHCreate]")

	var uplinks []string
	var remote_cmd string
	var somthingWentWrong string
	var err error
	var i int

	name := d.Get("name").(string)
	ports := d.Get("ports").(int)
	mtu := d.Get("mtu").(int)
	link_discovery_mode := d.Get("link_discovery_mode").(string)
	promiscuous_mode := d.Get("promiscuous_mode").(bool)
	mac_changes := d.Get("mac_changes").(bool)
	forged_transmits := d.Get("forged_transmits").(bool)
	somthingWentWrong = ""

	// Validate variables
	if ports == 0 {
		ports = 128
	}

	if mtu == 0 {
		mtu = 1500
	}

	if link_discovery_mode == "" {
		link_discovery_mode = "listen"
	}

	if link_discovery_mode != "down" && link_discovery_mode != "listen" &&
		link_discovery_mode != "advertise" && link_discovery_mode != "both" {
		return fmt.Errorf("link_discovery_mode must be one of down, listen, advertise or both")
	}

	uplinkCount, ok := d.Get("uplink.#").(int)
	if !ok {
		uplinkCount = 0
		uplinks[0] = ""
	}
	if uplinkCount > 32 {
		uplinkCount = 32
	}
	for i = 0; i < uplinkCount; i++ {
		prefix := fmt.Sprintf("uplink.%d.", i)

		if attr, ok := d.Get(prefix + "name").(string); ok && attr != "" {
			uplinks = append(uplinks, d.Get(prefix+"name").(string))
		}
	}

	//  Create vswitch
	remote_cmd = fmt.Sprintf("esxcli network vswitch standard add -P %d -v \"%s\"",
		ports, name)

	stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "create vswitch")
	if strings.Contains(stdout, "this name already exists") {
		d.SetId("")
		return fmt.Errorf("Failed to add vswitch: %s, it already exists\n", name)
	}
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to add vswitch: %s\n%s\n", stdout, err)
	}

	//  Set id
	d.SetId(name)

	err = vswitchUpdate(c, name, ports, mtu, uplinks, link_discovery_mode, promiscuous_mode, mac_changes, forged_transmits)
	if err != nil {
		somthingWentWrong = fmt.Sprintf("Failed to update vswitch: %s\n", err)
	}

	// Refresh
	ports, mtu, uplinks, link_discovery_mode, promiscuous_mode, mac_changes, forged_transmits, err = vswitchRead(c, name)
	if err != nil {
		d.SetId("")
		return nil
	}

	// Change uplinks (list) to map
	log.Printf("[resourceVSWITCHCreate] uplinks: %s\n", uplinks)
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

	if somthingWentWrong != "" {
		return fmt.Errorf(somthingWentWrong)
	}
	return nil
}
