package esxi

import (
	"log"
	"fmt"
	"errors"
)


func guestUPDATE(c *Config, vmid string, memsize int, numvcpus int, virthwver int,
	virtual_networks [4][3]string, virtual_disks [60][2]string, guest_shutdown_timeout int) error {
  log.Printf("[provider-esxi / guestUPDATE]\n")

  var err error

  //
	//   Power off guest if it's powered on.
	//
  currentpowerstate := guestPowerGetState(c, vmid)
	if currentpowerstate == "on" ||  currentpowerstate == "suspended" {
    _, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
	  if err != nil {
	  	return err
	  }
	}

  //
  //  make updates to vmx file
  //
  err = updateVmx_contents(c, vmid, false, memsize, numvcpus, virthwver, virtual_networks, virtual_disks)

	//  power on
	_, err = guestPowerOn(c, vmid)
	if err != nil {
		fmt.Println("Failed to power on.")
		return errors.New("Failed to power on.")
	}

  return err
}
