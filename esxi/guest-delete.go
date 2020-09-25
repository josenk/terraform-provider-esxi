package esxi

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGUESTDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[resourceGUESTDelete]")

	var remote_cmd, stdout string
	var err error

	vmid := d.Id()
	guest_shutdown_timeout := d.Get("guest_shutdown_timeout").(int)

	_, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
	if err != nil {
		return fmt.Errorf("Failed to power off: %s\n", err)
	}

	// remove storage from vmx so it doesn't get deleted by the vim-cmd destroy
	err = cleanStorageFromVmx(c, vmid)
	if err != nil {
		log.Printf("[resourceGUESTDelete] Failed clean storage from vmid: %s (to be deleted)\n", vmid)
	}

	time.Sleep(5 * time.Second)
	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/destroy")
	if err != nil {
		log.Printf("[resourceGUESTDelete] Failed destroy vmid: %s\n", stdout)
		return fmt.Errorf("Failed to destroy vm: %s\n", err)
	}

	d.SetId("")

	return nil
}
