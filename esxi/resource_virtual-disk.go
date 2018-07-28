package esxi

import (
  "fmt"
  "log"
  "strings"
  "strconv"
  "errors"
  "github.com/hashicorp/terraform/helper/schema"
  "math/rand"
	"time"
)

func resourceVIRTUALDISK() *schema.Resource {
  return &schema.Resource{
    Create: resourceVIRTUALDISKCreate,
    Read:   resourceVIRTUALDISKRead,
    Update: resourceVIRTUALDISKUpdate,
    Delete: resourceVIRTUALDISKDelete,
    Schema: map[string]*schema.Schema{
      "virtual_disk_disk_store": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_disk_store", nil),
          Description: "Disk Store.",
      },
      "virtual_disk_dir": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_dir", nil),
          Description: "Disk dir.",
      },
      "virtual_disk_name": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: true,
          Computed: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_name", nil),
          Description: "Virtual Disk Name. A random virtual disk name will be generated if nil.",
      },
      "virtual_disk_size": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_size", 1),
          Description: "Virtual Disk size in GB.",
      },
      "virtual_disk_type": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_type", "thin"),
          Description: "Virtual Disk type.  (thin, thick or eagerzeroedthick)",
      },
    },
  }
}

func resourceVIRTUALDISKCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

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
    log.Println("[provider-esxi] Error: " + err.Error())
    d.SetId("")
  }
  return nil
}

func resourceVIRTUALDISKRead(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name, virtual_disk_size, virtual_disk_type, err := virtualDiskREAD(c, d.Id())
  if err != nil {
    d.SetId("")
    return nil
  }

  d.Set("virtual_disk_disk_store", virtual_disk_disk_store)
  d.Set("virtual_disk_dir", virtual_disk_dir)
  d.Set("virtual_disk_name", virtual_disk_name)
  d.Set("virtual_disk_size", virtual_disk_size)

  if virtual_disk_type != "Unknown" {
    d.Set("virtual_disk_type", virtual_disk_type)
  }

  return nil
}

func resourceVIRTUALDISKUpdate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  if d.HasChange("virtual_disk_size") {
    _, _, _, curr_virtual_disk_size, _, err := virtualDiskREAD(c, d.Id())
    if err != nil {
      d.SetId("")
      return nil
    }

    virtual_disk_size       := d.Get("virtual_disk_size").(int)

    if curr_virtual_disk_size > virtual_disk_size {
      return errors.New("Not able to shrink virtual disk:" + d.Id())
    }

    err = growVirtualDisk(c, d.Id(), strconv.Itoa(virtual_disk_size))
		if err != nil {
			return errors.New("Unable to grow disk:" + d.Id())
		}
  }

  return nil
}

func resourceVIRTUALDISKDelete(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  virtual_disk_disk_store := d.Get("virtual_disk_disk_store").(string)
  virtual_disk_dir        := d.Get("virtual_disk_dir").(string)
  //  Not needed to delete the virtual disk.
  //virtual_disk_name       := d.Get("virtual_disk_name").(string)
  //virtual_disk_size       := d.Get("virtual_disk_size").(int)
  //virtual_disk_type       := d.Get("virtual_disk_type").(string)


  err := virtualDiskDELETE(c, d.Id(), virtual_disk_disk_store, virtual_disk_dir)

  // already deleted
  if strings.Contains(err.Error(), "Process exited with status 255") == true {
    log.Printf("[provider-esxi / resourceVIRTUALDISKDelete] Already deleted:%s", d.Id())
    d.SetId("")
    return nil
  }

  if err != nil {
    return err
  }

  d.SetId("")
  return nil
}
