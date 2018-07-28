package esxi

import (
	"fmt"
	"log"
)

//  Grow virtual Disk
func growVirtualDisk(c *Config, virtdisk_id string, virtdisk_size string) error {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / growVirtualDisk]\n")

	remote_cmd := fmt.Sprintf("/bin/vmkfstools -X %sG \"%s\"", virtdisk_size, virtdisk_id)
	_, err     := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "grow disk")

  return err
}
