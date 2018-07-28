package esxi

import (
	"fmt"
	"log"
  "strings"
  "strconv"
)


func virtualDiskREAD(c *Config, virtdisk_id string) (string, string, string, int, string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Println("[provider-esxi / virtualDiskREAD] Begin" )

  var virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name string
	var virtual_disk_type, flatSize string
	var virtual_disk_size int
	var flatSizei64 int64
	var s []string

	//  Split virtdisk_id into it's variables
	s = strings.Split(virtdisk_id, "/")
	log.Printf("[provider-esxi / virtualDiskREAD] len=%d cap=%d %v\n", len(s), cap(s), s)
	virtual_disk_disk_store = s[3]
	virtual_disk_dir = s[4]
	virtual_disk_name = s[5]

  // Test if virtual disk exists
  remote_cmd := fmt.Sprintf("test -s %s", virtdisk_id)
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

	remote_cmd = fmt.Sprintf("ls -l /vmfs/volumes/%s/%s/%s | awk '{print $5}'",
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
