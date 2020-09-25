package esxi

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVIRTUALDISKUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	//esxiConnInfo := ConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceVIRTUALDISKUpdate]")

	if d.HasChange("virtual_disk_size") {
		_, _, _, curr_virtual_disk_size, _, err := virtualDiskREAD(c, d.Id())
		if err != nil {
			d.SetId("")
			return fmt.Errorf("Failed to refresh virtual disk: %s\n", err)
		}

		virtual_disk_size := d.Get("virtual_disk_size").(int)

		if curr_virtual_disk_size > virtual_disk_size {
			return errors.New("Not able to shrink virtual disk:" + d.Id())
		}

		err = growVirtualDisk(c, d.Id(), strconv.Itoa(virtual_disk_size))
		if err != nil {
			return fmt.Errorf("Failed to grow virtual disk: %s\n", err)
		}
	}

	return nil
}
