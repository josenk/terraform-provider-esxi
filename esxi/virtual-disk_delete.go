package esxi

import (
	"fmt"
	"log"
)

func virtualDiskDELETE(c *Config, virtdisk_id string, virtual_disk_disk_store string, virtual_disk_dir string) error {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Println("[provider-esxi / virtualDiskDELETE] Begin" )

  var remote_cmd, stdout string
	var err error

  //  Destroy virtual disk.
  remote_cmd = fmt.Sprintf("/bin/vmkfstools -U %s", virtdisk_id)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy virtual disk")
  if err != nil {
		// todo more descriptive err message
    log.Printf("[provider-esxi / virtualDiskDELETE] Failed destroy virtual disk id: %s\n", stdout)
    return err
  }

  //  Delete dir if it's empty
	remote_cmd = fmt.Sprintf("ls -al /vmfs/volumes/%s/%s/ |wc -l", virtual_disk_disk_store, virtual_disk_dir)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Check if Storage dir is empty")
  if err != nil {
    log.Printf("[provider-esxi / virtualDiskDELETE] Unable to check if storage dir is empty: %s\n", stdout)
    return err
  }
  if stdout == "3" {
		{
			//  Delete empty dir.  Ignore stdout and errors.
			remote_cmd = fmt.Sprintf("rmdir /vmfs/volumes/%s/%s", virtual_disk_disk_store, virtual_disk_dir)
		  _, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "rmdir empty Storage dir")
			}
	}

  return nil
}
