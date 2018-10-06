package esxi

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
)

func resourceGUESTUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Printf("[guestUPDATE]\n")

	var virtual_networks [4][3]string
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
	lanAdaptersCount := d.Get("network_interfaces.#").(int)

	if lanAdaptersCount > 3 {
		lanAdaptersCount = 3
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
	}

	//  Validate virtual_disks
	virtualDiskCount := d.Get("virtual_disks.#").(int)
	if virtualDiskCount > 59 {
		virtualDiskCount = 59
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
			return err
		}
	}

	//
	//  make updates to vmx file
	//
	imemsize, _ := strconv.Atoi(memsize)
	inumvcpus, _ := strconv.Atoi(numvcpus)
	ivirthwver, _ := strconv.Atoi(virthwver)
	err = updateVmx_contents(c, vmid, false, imemsize, inumvcpus, ivirthwver, guestos, virtual_networks, virtual_disks)
	if err != nil {
		fmt.Println("Failed to update VMX file.")
		return errors.New("Failed to update VMX file.")
	}

	//
	//  Grow boot disk to boot_disk_size
	//
	boot_disk_vmdkPATH, _ := getBootDiskPath(c, vmid)

	err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
	if err != nil {
		return errors.New("Failed to grow boot disk.")
	}

	//  power on
	_, err = guestPowerOn(c, vmid)
	if err != nil {
		fmt.Println("Failed to power on.")
		return errors.New("Failed to power on.")
	}

	return err
}
