package esxi

import (
	"fmt"
	"log"
)


func guestDELETE(c *Config, vmid string, guest_shutdown_timeout int) error {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}

	_, err := guestPowerOff(c, vmid, guest_shutdown_timeout)
	if err != nil {
		return err
	}

  remote_cmd := fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)
	stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/destroy")
	if err != nil {
		// todo more descriptive err message
		log.Printf("[provider-esxi] Failed destroy vmid: %s\n", stdout)
		return err
	}
  return err
}
