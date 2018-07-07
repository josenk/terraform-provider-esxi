package esxi

import (
	"log"
)


func guestUPDATE(c *Config, vmid string, memsize string, numvcpus string, virthwver string,
	virtual_networks [4][3]string) error {
  log.Printf("[provider-esxi / guestUPDATE]")

  var err error

  savedpowerstate := guestPowerGetState(c, vmid)
	  if savedpowerstate == "on" ||  savedpowerstate == "suspended" {
    _, err = guestPowerOff(c, vmid)
	  if err != nil {
	  	return err
	  }
	}

  //
  //  make updates to vmx file
  //
  err = updateVmx_contents(c, vmid, false, memsize, numvcpus, virthwver, virtual_networks)

  return err
}
