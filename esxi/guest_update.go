package esxi

import (
//	"fmt"
	"log"
//  "strconv"
//  "strings"
)


func guestUPDATE(c *Config, vmid string, memsize string, numvcpus string) error {
  log.Printf("[provider-esxi / guestUPDATE]")

  _, err := guestPowerOff(c, vmid)
	if err != nil {
		return err
	}

  //
  //  make updates to vmx file
  //
  err = updateVmx_contents(c, vmid, memsize, numvcpus)

  _, err = guestPowerOn(c, vmid)
	if err != nil {
		return err
	}

  return err
}
