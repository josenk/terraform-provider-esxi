package esxi

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"github.com/hashicorp/terraform/helper/schema"
)


func resourceVIRTUALDISKCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
	//esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
  log.Println("[resourceVIRTUALDISKCreate]" )

  virtual_disk_disk_store := d.Get("virtual_disk_disk_store").(string)
  virtual_disk_dir        := d.Get("virtual_disk_dir").(string)
  virtual_disk_name       := d.Get("virtual_disk_name").(string)
  virtual_disk_size       := d.Get("virtual_disk_size").(int)
  virtual_disk_type       := d.Get("virtual_disk_type").(string)

  if virtual_disk_name == "" {
    rand.Seed(time.Now().UnixNano())

	  const digits = "0123456789ABCDEF"
	  name := make([]byte, 10)
    for i := range name {
	    name[i] = digits[rand.Intn(len(digits))]
	  }

    virtual_disk_name = fmt.Sprintf("vdisk_%s.vmdk", name)
  }

  //
  //  Validate virtual_disk_name
  //

  // todo,  check invalid chars (quotes, slash, period, comma)
  // todo,  must end with .vmdk


  virtdisk_id,err := virtualDiskCREATE(c, virtual_disk_disk_store, virtual_disk_dir,
    virtual_disk_name, virtual_disk_size, virtual_disk_type)
  if err == nil {
    d.SetId(virtdisk_id)
  } else {
    log.Println("[resourceVIRTUALDISKCreate] Error: " + err.Error())
    d.SetId("")
		return fmt.Errorf("Failed to create virtual Disk :%s\nError:%s", virtual_disk_name, err.Error())
  }

  return nil
}
