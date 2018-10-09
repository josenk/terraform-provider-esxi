package esxi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceVIRTUALDISKDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceVIRTUALDISKDelete]")

	var remote_cmd, stdout string
	var err error

	virtdisk_id := d.Id()
	virtual_disk_disk_store := d.Get("virtual_disk_disk_store").(string)
	virtual_disk_dir := d.Get("virtual_disk_dir").(string)

	//  Destroy virtual disk.
	remote_cmd = fmt.Sprintf("/bin/vmkfstools -U %s", virtdisk_id)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy virtual disk")
	if err != nil {
		if strings.Contains(err.Error(), "Process exited with status 255") == true {
			log.Printf("[resourceVIRTUALDISKDelete] Already deleted:%s", virtdisk_id)
		} else {
			// todo more descriptive err message
			log.Printf("[resourceVIRTUALDISKDelete] Failed destroy virtual disk id: %s\n", stdout)
			return err
		}
	}

	//  Delete dir if it's empty
	remote_cmd = fmt.Sprintf("ls -al \"/vmfs/volumes/%s/%s/\" |wc -l", virtual_disk_disk_store, virtual_disk_dir)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Check if Storage dir is empty")
	if stdout == "3" {
		{
			//  Delete empty dir.  Ignore stdout and errors.
			remote_cmd = fmt.Sprintf("rmdir \"/vmfs/volumes/%s/%s\"", virtual_disk_disk_store, virtual_disk_dir)
			_, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "rmdir empty Storage dir")
		}
	}

	d.SetId("")
	return nil
}
