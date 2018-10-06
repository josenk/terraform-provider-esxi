package esxi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
)

func resourceVIRTUALDISKUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	//esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceVIRTUALDISKUpdate]")

	if d.HasChange("virtual_disk_size") {
		_, _, _, curr_virtual_disk_size, _, err := virtualDiskREAD(c, d.Id())
		if err != nil {
			d.SetId("")
			return nil
		}

		virtual_disk_size := d.Get("virtual_disk_size").(int)

		if curr_virtual_disk_size > virtual_disk_size {
			return errors.New("Not able to shrink virtual disk:" + d.Id())
		}

		err = growVirtualDisk(c, d.Id(), strconv.Itoa(virtual_disk_size))
		if err != nil {
			return errors.New("Failed to grow disk:" + d.Id())
		}
	}

	return nil
}
