package esxi

import (
  "github.com/hashicorp/terraform/helper/schema"
)

func resourceVIRTUALDISK() *schema.Resource {
  return &schema.Resource{
    Create: resourceVIRTUALDISKCreate,
    Read:   resourceVIRTUALDISKRead,
    Update: resourceVIRTUALDISKUpdate,
    Delete: resourceVIRTUALDISKDelete,
    Importer: &schema.ResourceImporter{
			State: resourceVIRTUALDISKImport,
    },
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
          Computed: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_size", 1),
          Description: "Virtual Disk size in GB.",
      },
      "virtual_disk_type": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("virtual_disk_type", "thin"),
          Description: "Virtual Disk type.  (thin, zeroedthick or eagerzeroedthick)",
      },
    },
  }
}
