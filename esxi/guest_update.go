package esxi

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGUESTUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Printf("[resourceGUESTUpdate]\n")

	var virtual_networks [10][4]string
	var virtual_disks [60][2]string
	var i int
	var err error

	vmid := d.Id()
	memsize := d.Get("memsize").(string)
	numvcpus := d.Get("numvcpus").(string)
	boot_disk_size := d.Get("boot_disk_size").(string)
	virthwver := d.Get("virthwver").(string)
	guestos := d.Get("guestos").(string)
	guest_shutdown_timeout := d.Get("guest_shutdown_timeout").(int)
	notes := d.Get("notes").(string)
	lanAdaptersCount := d.Get("network_interfaces.#").(int)
	power := d.Get("power").(string)

	guestinfo, ok := d.Get("guestinfo").(map[string]interface{})
	if !ok {
		return errors.New("guestinfo is wrong type")
	}

	if lanAdaptersCount > 10 {
		lanAdaptersCount = 10
	}
	for i := 0; i < lanAdaptersCount; i++ {
		prefix := fmt.Sprintf("network_interfaces.%d.", i)

		if attr, ok := d.Get(prefix + "virtual_network").(string); ok && attr != "" {
			virtual_networks[i][0] = d.Get(prefix + "virtual_network").(string)
		}
		if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
			virtual_networks[i][1] = d.Get(prefix + "mac_address").(string)
		}
		if attr, ok := d.Get(prefix + "nic_type").(string); ok && attr != "" {
			virtual_networks[i][2] = d.Get(prefix + "nic_type").(string)
		}
		if attr, ok := d.Get(prefix + "ovf_network").(string); ok && attr != "" {
			virtual_networks[i][3] = d.Get(prefix + "ovf_network").(string)
		}
	}

	//  Validate virtual_disks
	virtualDiskCount := d.Get("virtual_disks.#").(int)
	if virtualDiskCount > 59 {
		virtualDiskCount = 59
	}

	// Validate guestOS
	if validateGuestOsType(guestos) == false {
		return errors.New("Error: invalid guestos.  see https://github.com/josenk/vagrant-vmware-esxi/wiki/VMware-ESXi-6.5-guestOS-types")
	}

	for i = 0; i < virtualDiskCount; i++ {
		prefix := fmt.Sprintf("virtual_disks.%d.", i)

		if attr, ok := d.Get(prefix + "virtual_disk_id").(string); ok && attr != "" {
			virtual_disks[i][0] = d.Get(prefix + "virtual_disk_id").(string)
		}

		if attr, ok := d.Get(prefix + "slot").(string); ok && attr != "" {
			// todo validate slots are in format "0-3:0-15"
			virtual_disks[i][1] = d.Get(prefix + "slot").(string)
		}
	}

	//
	//   Power off guest if it's powered on.
	//
	currentpowerstate := guestPowerGetState(c, vmid)
	if currentpowerstate == "on" || currentpowerstate == "suspended" {
		_, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
		if err != nil {
			return fmt.Errorf("Failed to power off: %s\n", err)
		}
	}

	//
	//  make updates to vmx file
	//
	imemsize, _ := strconv.Atoi(memsize)
	inumvcpus, _ := strconv.Atoi(numvcpus)
	ivirthwver, _ := strconv.Atoi(virthwver)
	err = updateVmx_contents(c, vmid, false, imemsize, inumvcpus, ivirthwver, guestos, virtual_networks, virtual_disks, notes, guestinfo)
	if err != nil {
		fmt.Println("Failed to update vmx file.")
		return fmt.Errorf("Failed to update vmx file: %s\n", err)
	}

	//
	//  Grow boot disk to boot_disk_size
	//
	boot_disk_vmdkPATH, _ := getBootDiskPath(c, vmid)

	err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
	if err != nil {
		return fmt.Errorf("Failed to grow virtual disk: %s\n", err)
	}

	//  power on
	if power == "on" {
		_, err = guestPowerOn(c, vmid)
		if err != nil {
			fmt.Println("Failed to power on.")
			return fmt.Errorf("Failed to power on: %s\n", err)
		}
	}

	return resourceGUESTRead(d, m)
}
