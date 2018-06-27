package esxi

import (
	"log"
)


func guestUPDATE(c *Config, vmid string, memsize string, numvcpus string,
	virtual_networks [4][3]string) error {
  log.Printf("[provider-esxi / guestUPDATE]")

  _, err := guestPowerOff(c, vmid)
	if err != nil {
		return err
	}

  //
  //  make updates to vmx file
  //
  err = updateVmx_contents(c, vmid, false, memsize, numvcpus, virtual_networks)

  _, err = guestPowerOn(c, vmid)
	if err != nil {
		return err
	}

  return err
}
