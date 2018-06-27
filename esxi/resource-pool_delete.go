package esxi

import (
	"fmt"
	"log"
)


func resourcePoolDELETE(c *Config, pool_id string) error {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Println("[provider-esxi / resourcePoolDELETE] Begin" )


  remote_cmd := fmt.Sprintf("vim-cmd hostsvc/rsrc/destroy %s", pool_id)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy resource pool")
  if err != nil {
    log.Printf("[provider-esxi / resourcePoolDELETE] Failed destroy resource pool id: %s", stdout)
    return err
  }
  return err
}
