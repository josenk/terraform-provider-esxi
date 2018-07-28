package esxi

import (
	"fmt"
	"log"
	"errors"
)


func virtualDiskCREATE(c *Config, virtual_disk_disk_store string, virtual_disk_dir string,
	virtual_disk_name string, virtual_disk_size int, virtual_disk_type string) (string, error) {

  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
	log.Println("[provider-esxi / virtualDiskCREATE]" )
	var virtdisk_id, remote_cmd string
	var err error

  //
	//  Validate disk store exists
	//
  remote_cmd = fmt.Sprintf("ls -d /vmfs/volumes/%s", virtual_disk_disk_store)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate disk store exists")
	if err != nil {
    return "", errors.New("virtual_disk_disk_store does not exist.")
  }

	//
	//  Create dir if required
  //
	remote_cmd = fmt.Sprintf("mkdir -p /vmfs/volumes/%s/%s", virtual_disk_disk_store, virtual_disk_dir)
	_, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create virtual disk dir")

	remote_cmd = fmt.Sprintf("ls -d /vmfs/volumes/%s/%s", virtual_disk_disk_store, virtual_disk_dir)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate dir exists")
	if err != nil {
		return "", errors.New("Unable to create virtual_disk directory.")
	}

  //
	//  Create virtual disk
	//
	virtdisk_id = fmt.Sprintf("/vmfs/volumes/%s/%s/%s", virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name)

	remote_cmd = fmt.Sprintf("/bin/vmkfstools -c %dG -d %s %s", virtual_disk_size,
		virtual_disk_type, virtdisk_id)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Create virtual_disk")
	if err != nil {
		return "", errors.New("Unable to create virtual_disk")
	}

  return virtdisk_id, err
}
