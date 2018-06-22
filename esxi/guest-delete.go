package esxi

import (
	"fmt"
	"log"
)


func guestDELETE(c *Config, vmid string) error {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}

	_, err := guestPowerOff(c, vmid)
	if err != nil {
		return err
	}

  remote_cmd := fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)
	stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/destroy")
	if err != nil {
		log.Printf("[provider-esxi] Failed destroy vmid: %s", stdout)
		return err
	}
  return err
}
