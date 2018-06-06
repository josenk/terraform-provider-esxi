package esxi

import (
	"fmt"
	"log"
)


func GuestDelete(c *Config, vmid string) error {

  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  remote_cmd := fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)
	stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get vmid")
	if err != nil {
		log.Printf("[provider-esxi] Failed destroy vmid: %s", stdout)
		return err
	}
  return err
}
