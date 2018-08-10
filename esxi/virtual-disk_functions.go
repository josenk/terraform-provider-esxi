package esxi

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"errors"
)


//
//  Create virtual disk
//
func virtualDiskCREATE(c *Config, virtual_disk_disk_store string, virtual_disk_dir string,
	virtual_disk_name string, virtual_disk_size int, virtual_disk_type string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualDiskCREATE]" )

	var virtdisk_id, remote_cmd string
	var err error

  //
	//  Validate disk store exists
	//
  remote_cmd = fmt.Sprintf("ls -d \"/vmfs/volumes/%s\"", virtual_disk_disk_store)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate disk store exists")
	if err != nil {
    return "", errors.New("virtual_disk_disk_store does not exist.")
  }

	//
	//  Create dir if required
  //
	remote_cmd = fmt.Sprintf("mkdir -p \"/vmfs/volumes/%s/%s\"", virtual_disk_disk_store, virtual_disk_dir)
	_, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create virtual disk dir")

	remote_cmd = fmt.Sprintf("ls -d \"/vmfs/volumes/%s/%s\"", virtual_disk_disk_store, virtual_disk_dir)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate dir exists")
	if err != nil {
		return "", errors.New("Unable to create virtual_disk directory.")
	}

  //
	//  virtdisk_id is just the full path name.
	//
	virtdisk_id = fmt.Sprintf("/vmfs/volumes/%s/%s/%s", virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name)

	//
	//  Validate if it exists already
	//
	remote_cmd = fmt.Sprintf("ls -l \"%s\"", virtdisk_id)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate disk store exists")
	if err == nil {
		log.Println("[virtualDiskCREATE]  Already exists." )
		return virtdisk_id, err
	}

	remote_cmd = fmt.Sprintf("/bin/vmkfstools -c %dG -d %s \"%s\"", virtual_disk_size,
		virtual_disk_type, virtdisk_id)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Create virtual_disk")
	if err != nil {
		return "", errors.New("Unable to create virtual_disk")
	}

  return virtdisk_id, err
}

//
//  Grow virtual Disk
//
func growVirtualDisk(c *Config, virtdisk_id string, virtdisk_size string) error {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
  log.Printf("[growVirtualDisk]\n")

	var newDiskSize int

	_, _, _, currentDiskSize, _, err := virtualDiskREAD(c, virtdisk_id)

	newDiskSize, _ = strconv.Atoi(virtdisk_size)

  log.Printf("[provider-esxi] currentDiskSize:%d new_size:%d fullPATH: %s\n", currentDiskSize, newDiskSize, virtdisk_id)

  if currentDiskSize < newDiskSize {
	  remote_cmd := fmt.Sprintf("/bin/vmkfstools -X %dG \"%s\"", newDiskSize, virtdisk_id)
	  _, err     := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "grow disk")
		if err != nil {
			return err
		}
	}

  return err
}


//
//  Read virtual Disk details
//
func virtualDiskREAD(c *Config, virtdisk_id string) (string, string, string, int, string, error) {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
  log.Println("[virtualDiskREAD] Begin" )

  var virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name string
	var virtual_disk_type, flatSize string
	var virtual_disk_size int
	var flatSizei64 int64
	var s []string

	//  Split virtdisk_id into it's variables
	s = strings.Split(virtdisk_id, "/")
	log.Printf("[virtualDiskREAD] len=%d cap=%d %v\n", len(s), cap(s), s)
	virtual_disk_disk_store = s[3]
	virtual_disk_dir = s[4]
	virtual_disk_name = s[5]

  // Test if virtual disk exists
  remote_cmd := fmt.Sprintf("test -s \"%s\"", virtdisk_id)
  _, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "test if virtual disk exists")
	if err != nil {
		return "","","",0,"", err
	}

  //  Get virtual disk flat size
	s = strings.Split(virtual_disk_name, ".")
	if len(s) < 2 {
		return "","","",0,"", err
	}
	virtual_disk_nameFlat := fmt.Sprintf("%s-flat.%s", s[0], s[1])

	remote_cmd = fmt.Sprintf("ls -l \"/vmfs/volumes/%s/%s/%s\" | awk '{print $5}'",
		virtual_disk_disk_store, virtual_disk_dir, virtual_disk_nameFlat)
	flatSize, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get size")
	if err != nil {
		return "","","",0,"", err
	}
	flatSizei64, _ = strconv.ParseInt(flatSize, 10, 64)
	virtual_disk_size = int(flatSizei64 / 1024 / 1024 / 1024)

	// Determine virtual disk type
	virtual_disk_type = "Unknown"  // todo

  // Return results

  return virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name, virtual_disk_size, virtual_disk_type, err
}
