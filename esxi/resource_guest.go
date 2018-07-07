package esxi

import (
  "fmt"
  "errors"
  "github.com/hashicorp/terraform/helper/schema"
  "strconv"
  "log"
  "strings"
)

func resourceGUEST() *schema.Resource {
  return &schema.Resource{
    Create: resourceGUESTCreate,
    Read:   resourceGUESTRead,
    Update: resourceGUESTUpdate,
    Delete: resourceGUESTDelete,
    Schema: map[string]*schema.Schema{
      "clone_from_vm": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("clone_from_vm", nil),
          Description: "Source vm path on esxi host to clone.",
      },
      "ovf_source": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("ovf_source", nil),
          Description: "Local path to source ovf files.",
      },
      "disk_store": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          DefaultFunc: schema.EnvDefaultFunc("disk_store", "Least Used"),
          Description: "esxi diskstore for boot disk.",
      },
      "resource_pool_name": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("resource_pool_name", "/"),
          Description: "Use resource pool.",
      },
      "guest_name": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("guest_name", "vm-example"),
          Description: "esxi guest name.",
      },
      "boot_disk_type": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("boot_disk_type", nil),
          Description: "Guest boot disk type. thin, thick, eagerzeroedthick",
      },
      "boot_disk_size": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("boot_disk_size", nil),
          Description: "Guest boot disk size. Will expand boot disk to this size.",
      },
      //"guest_storage": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("guest_storage", nil),
      //    Description: "Guest guest additional storage.",
      //},
      "memsize": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("memsize", nil),
          Description: "Guest guest memory size.",
      },
      "numvcpus": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("numvcpus", nil),
          Description: "Guest guest number of virtual cpus.",
      },
      "virthwver": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("virthwver", nil),
          Description: "Guest Virtual HW version.",
      },
      "network_interfaces": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"virtual_network": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"mac_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"nic_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
					},
				},
      },
      "power": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("power", nil),
          Description: "Guest powered on.",
      },
    },
  }
}

func resourceGUESTCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  var virtual_networks [4][3]string

  clone_from_vm      := d.Get("clone_from_vm").(string)
  ovf_source         := d.Get("ovf_source").(string)
  disk_store         := d.Get("disk_store").(string)
  resource_pool_name := d.Get("resource_pool_name").(string)
  guest_name         := d.Get("guest_name").(string)
  boot_disk_type     := d.Get("boot_disk_type").(string)
  boot_disk_size     := d.Get("boot_disk_size").(string)
  //guest_storage      := d.Get("guest_storage").(string)
  memsize            := d.Get("memsize").(string)
  numvcpus           := d.Get("numvcpus").(string)
  virthwver          := d.Get("virthwver").(string)
  power              := d.Get("power").(string)

  // Validations
  var src_path string
  var tmpint int

  if resource_pool_name == "ha-root-pool" {
    resource_pool_name = "/"
  }

  if clone_from_vm != "" {
    src_path = fmt.Sprintf("vi://%s:%s@%s/%s", c.Esxi_username, c.Esxi_password, c.Esxi_hostname, clone_from_vm)
    fmt.Println("[Terraform-provider-esxi]   ")
  } else if ovf_source != "" {
    src_path = ovf_source
  } else {
    src_path = "none"
  }

  //  Validate memsize
  if _, err := strconv.Atoi(memsize); err != nil && memsize != "" {
    return errors.New("Error: memsize must be an integer")
  }
  tmpint,_ = strconv.Atoi(memsize)
  if (tmpint < 128 || tmpint > 6128) && memsize != "" {
    return errors.New("Error: memsize must be > 128 and < 6128000")
  }

  //  Validate number of virt cpus.
  if _, err := strconv.Atoi(numvcpus); err != nil && numvcpus != "" {
    return errors.New("Error: numvcpus must be an integer")
  }
  tmpint,_ = strconv.Atoi(numvcpus)
  if (tmpint < 1 || tmpint >128) && numvcpus != "" {
    return errors.New("Error: numvcpus must be an > 0 and < 128")
  }

  //  Validate number of virthwver.
  if _, err := strconv.Atoi(virthwver); err != nil && virthwver != "" {
    return errors.New("Error: virthwver must be an integer")
  }
  if virthwver != "" {
    tmpint,_ = strconv.Atoi(virthwver)
    switch tmpint {
    case 4,7,8,9,10,11,12,13,14:
      // virthwver check passes.
    default:
      return errors.New("Error: virthwver must be 4,7,8,9,10,11,12,13 or 14")
    }
  }

  // Validate boot_disk_type
  if boot_disk_type == "" {
    boot_disk_type = "thin"
  }
  if boot_disk_type != "thin" && boot_disk_type != "thick" && boot_disk_type != "eagerzeroedthick" {
    return errors.New("Error: boot_disk_type must be thin, thick or eagerzeroedthick")
  }

  //  Validate boot_disk_size.
  if _, err := strconv.Atoi(boot_disk_size); err != nil && boot_disk_size != "" {
    return errors.New("Error: boot_disk_size must be an integer")
  }
  tmpint,_ = strconv.Atoi(boot_disk_size)
  if (tmpint < 1 || tmpint >62000) && boot_disk_size != "" {
    return errors.New("Error: boot_disk_size must be an > 1 and < 62000")
  }

  adaptersCount := d.Get("network_interfaces.#").(int)
  if adaptersCount > 3 {
    adaptersCount = 3
  }
  for i := 0; i < adaptersCount; i++ {
    prefix := fmt.Sprintf("network_interfaces.%d.", i)

    if attr, ok := d.Get(prefix + "virtual_network").(string); ok && attr != "" {
			virtual_networks[i][0] = d.Get(prefix + "virtual_network").(string)
    }

    if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
      virtual_networks[i][1] = d.Get(prefix + "mac_address").(string)
    }

    if attr, ok := d.Get(prefix + "nic_type").(string); ok && attr != "" {
      if strings.Contains("vlance flexible e1000 e1000e vmxnet vmxnet2 vmxnet3",
        d.Get(prefix + "nic_type").(string)) == true {

        virtual_networks[i][2] = d.Get(prefix + "nic_type").(string)

      } else {

        return errors.New("Error: Unsupported nic_type. (vlance flexible e1000 e1000e vmxnet vmxnet2 vmxnet3)")

      }
    }
  }


  vmid, err := guestCREATE(c, guest_name, disk_store, src_path, resource_pool_name,
     memsize, numvcpus, virthwver, boot_disk_type, boot_disk_size, virtual_networks)
  if err != nil {
    tmpint,_ = strconv.Atoi(vmid)
    if tmpint > 0 {
      d.SetId(vmid)
      fmt.Println("Error: There was an error while creating guest.")
      return errors.New(vmid)
    } else {
      fmt.Println("Error: Unable to create guest.")
      return errors.New(vmid)
    }
  }
  d.SetId(vmid)

  // Do Power state
  if power == "on" || power == "true" || power == "" {
    _, err = guestPowerOn(c, vmid)
    if err != nil {
      fmt.Println("Failed to power on.")
      return errors.New("Failed to power on.")
    }
  }

  return nil
}

func resourceGUESTRead(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  guest_name, disk_store, resource_pool_name, memsize, numvcpus, virthwver, virtual_networks, err := guestREAD(c, d.Id())

  if err != nil {
    d.SetId("")
  }
  d.Set("disk_store",disk_store)
  d.Set("resource_pool_name",resource_pool_name)
  d.Set("guest_name",guest_name)
  if d.Get("memsize").(string) != "" {
    d.Set("memsize",memsize)
  }
  if d.Get("numvcpus").(string) != "" {
    d.Set("numvcpus",numvcpus)
  }
  if d.Get("virthwver").(string) != "" {
    d.Set("virthwver",virthwver)
  }


  // Do network interfaces
  log.Printf("virtual_networks: %q", virtual_networks)
  nics := make([]map[string]interface{}, 0, 1)

	for nic := 0; nic < 3; nic++ {
    if virtual_networks[nic][0] != "" {

      prefix := fmt.Sprintf("network_interfaces.%d.", nic)

		  out := make(map[string]interface{})
		  out["virtual_network"] = virtual_networks[nic][0]
      if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
		    out["mac_address"]     = virtual_networks[nic][1]
      } else {
        out["mac_address"]     = nil
      }
      if attr, ok := d.Get(prefix + "nic_type").(string); ok && attr != "" {
		    out["nic_type"]        = virtual_networks[nic][2]
      } else {
        out["nic_type"]        = nil
      }

		  nics = append(nics, out)
    }
	}

  d.Set("network_interfaces", nics)

  // Do power state
  if d.Get("power").(string) != "" {
    d.Set("power", guestPowerGetState(c, d.Id()))
  }

  return nil
}

func resourceGUESTUpdate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  var virtual_networks [4][3]string

  memsize      := d.Get("memsize").(string)
  numvcpus     := d.Get("numvcpus").(string)
  virthwver    := d.Get("virthwver").(string)
  power        := d.Get("power").(string)

  adaptersCount := d.Get("network_interfaces.#").(int)
  if adaptersCount > 3 {
    adaptersCount = 3
  }
  for i := 0; i < adaptersCount; i++ {
    prefix := fmt.Sprintf("network_interfaces.%d.", i)

    if attr, ok := d.Get(prefix + "virtual_network").(string); ok && attr != "" {
      virtual_networks[i][0] = d.Get(prefix + "virtual_network").(string)
    }
    if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
      virtual_networks[i][1] = d.Get(prefix + "mac_address").(string)
    }
    if attr, ok := d.Get(prefix + "nic_type").(string); ok && attr != "" {
      virtual_networks[i][2] = d.Get(prefix + "nic_type").(string)
    }
    fmt.Printf("virtual_network:%s", virtual_networks[i][0])
    fmt.Printf("mac_address:%s", virtual_networks[i][1])
    fmt.Printf("nic_type: %s", virtual_networks[i][2])
  }

  err := guestUPDATE(c, d.Id(), memsize, numvcpus, virthwver, virtual_networks)

  // Do power state
  if power == "on" || power == "true" {
    _, err = guestPowerOn(c, d.Id())
    if err != nil {
      fmt.Println("Failed to power on.")
      return errors.New("Failed to power on.")
    }
  } else if power == "off" || power == "false" {
    _, err = guestPowerOff(c, d.Id())
    if err != nil {
      fmt.Println("Failed to power off.")
      return errors.New("Failed to power off.")
    }
  }

  return err
}

func resourceGUESTDelete(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  err := guestDELETE(c, d.Id())
  if err != nil {
    return err
  }
  d.SetId("")
  return nil
}
