package esxi

import (
  "fmt"
  "errors"
  "github.com/hashicorp/terraform/helper/schema"
  "github.com/hashicorp/terraform/helper/validation"
  "strconv"
  "log"
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
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("disk_store", "Least Used"),
          Description: "esxi diskstore for boot disk.",
      },
      "resource_pool_name": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: true,
          Computed: true,
          Description: "Resource pool name to place guest.",
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
          Computed: true,
          DefaultFunc: schema.EnvDefaultFunc("boot_disk_size", nil),
          Description: "Guest boot disk size. Will expand boot disk to this size.",
      },
      "memsize": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Guest guest memory size.",
      },
      "numvcpus": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Guest guest number of virtual cpus.",
      },
      "virthwver": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Guest Virtual HW version.",
      },
      "guestos": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Guest OS type.",
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
              Computed: true,
						},
						"nic_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
              Computed: true,
						},
					},
				},
      },
      "power": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Guest power state.",
          DefaultFunc: schema.EnvDefaultFunc("power", "on"),
      },
      //  Calculated only, you cannot overwrite this.
      "ip_address": &schema.Schema{
        Type:     schema.TypeString,
        Computed: true,
        Description: "The IP address reported by VMware tools.",
      },
      "guest_startup_timeout": {
        Type:        schema.TypeInt,
        Optional:    true,
        Default:     60,
        Description: "The amount of guest uptime, in seconds, to wait for an available IP address on this virtual machine.",
        ValidateFunc: validation.IntBetween(1, 600),
      },
      "guest_shutdown_timeout": {
        Type:        schema.TypeInt,
        Optional:    true,
        Default:     20,
        Description: "The amount of time, in seconds, to wait for a graceful shutdown before doing a forced power off.",
        ValidateFunc: validation.IntBetween(0, 600),
      },
      "virtual_disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"virtual_disk_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"slot": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
              Computed: true,
              Description: "SCSI_Ctrl:SCSI_id.    Range  '0:1' to '0:15'.   SCSI_id 7 is not allowed.",
						},
					},
				},
      },
    },
  }
}

func resourceGUESTCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  var virtual_networks [4][3]string
  var virtual_disks    [60][2]string
  var src_path string
  var tmpint, i int

  clone_from_vm      := d.Get("clone_from_vm").(string)
  ovf_source         := d.Get("ovf_source").(string)
  disk_store         := d.Get("disk_store").(string)
  resource_pool_name := d.Get("resource_pool_name").(string)
  guest_name         := d.Get("guest_name").(string)
  boot_disk_type     := d.Get("boot_disk_type").(string)
  boot_disk_size     := d.Get("boot_disk_size").(string)
  memsize            := d.Get("memsize").(string)
  numvcpus           := d.Get("numvcpus").(string)
  virthwver          := d.Get("virthwver").(string)
  guestos            := d.Get("guestos").(string)
  guest_shutdown_timeout  := d.Get("guest_shutdown_timeout").(int)
  guest_startup_timeout  := d.Get("guest_startup_timeout").(int)

  // Validations
  if resource_pool_name == "ha-root-pool" {
    resource_pool_name = "/"
  }

  if clone_from_vm != "" {
    src_path = fmt.Sprintf("vi://%s:%s@%s/%s", c.esxiUserName, c.esxiPassword, c.esxiHostName, clone_from_vm)
  } else if ovf_source != "" {
    src_path = ovf_source
  } else {
    src_path = "none"
  }

  //  Validate number of virthwver.
  // todo
  //switch virthwver {
  //case 0,4,7,8,9,10,11,12,13,14:
  //  // virthwver check passes.
  //default:
  //  return errors.New("Error: virthwver must be 4,7,8,9,10,11,12,13 or 14")
  //}


  //  Validate guestos
  if validateGuestOsType(guestos) == false {
    return errors.New("Error: invalid guestos.  see https://github.com/josenk/vagrant-vmware-esxi/wiki/VMware-ESXi-6.5-guestOS-types")
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

  //  Validate lan adapters
  lanAdaptersCount := d.Get("network_interfaces.#").(int)
  if lanAdaptersCount > 3 {
    lanAdaptersCount = 3
  }
  for i = 0; i < lanAdaptersCount; i++ {
    prefix := fmt.Sprintf("network_interfaces.%d.", i)

    if attr, ok := d.Get(prefix + "virtual_network").(string); ok && attr != "" {
			virtual_networks[i][0] = d.Get(prefix + "virtual_network").(string)
    }

    if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
      virtual_networks[i][1] = d.Get(prefix + "mac_address").(string)
    }

    if attr, ok := d.Get(prefix + "nic_type").(string); ok && attr != "" {
      virtual_networks[i][2] = d.Get(prefix + "nic_type").(string)
      //  Validate nictype
      if validateNICType(virtual_networks[i][2]) == false {
        errMSG := fmt.Sprintf("Error: invalid nic_type. %s\nMust be vlance flexible e1000 e1000e vmxnet vmxnet2 or vmxnet3", virtual_networks[i][2])
        return errors.New(errMSG)
      }
    }
  }

  //  Validate virtual_disks
  virtualDiskCount := d.Get("virtual_disks.#").(int)
  if virtualDiskCount > 59 {
    virtualDiskCount = 59
  }
  for i = 0; i < virtualDiskCount; i++ {
    prefix := fmt.Sprintf("virtual_disks.%d.", i)

    if attr, ok := d.Get(prefix + "virtual_disk_id").(string); ok && attr != "" {
			virtual_disks[i][0] = d.Get(prefix + "virtual_disk_id").(string)
    }

    if attr, ok := d.Get(prefix + "slot").(string); ok && attr != "" {
      virtual_disks[i][1] = d.Get(prefix + "slot").(string)
      result := validateVirtualDiskSlot(virtual_disks[i][1])
      if result != "ok" {
        return errors.New(result)
      }
    }
  }

  vmid, err := guestCREATE(c, guest_name, disk_store, src_path, resource_pool_name, memsize,
     numvcpus, virthwver, guestos, boot_disk_type, boot_disk_size, virtual_networks, virtual_disks, guest_shutdown_timeout)
  if err != nil {
    tmpint,_ = strconv.Atoi(vmid)
    if tmpint > 0 {
      d.SetId(vmid)
    }
    return err
  }

  //  set vmid
  d.SetId(vmid)

  _, err = guestPowerOn(c, vmid)
  if err != nil {
    return errors.New("Failed to power on.")
  }
  d.Set("power", "on")

  // Refresh
  guest_name, disk_store, disk_size, resource_pool_name, memsize, numvcpus, virthwver, guestos, ip_address, virtual_networks, power, err := guestREAD(c, d.Id(), guest_startup_timeout)
  if err != nil {
    d.SetId("")
    return nil
  }

  d.Set("guest_name",guest_name)
  d.Set("disk_store",disk_store)
  d.Set("disk_size",disk_size)
  d.Set("resource_pool_name",resource_pool_name)
  d.Set("memsize",memsize)
  d.Set("numvcpus",numvcpus)
  d.Set("virthwver",virthwver)
  d.Set("guestos",guestos)
  d.Set("ip_address", ip_address)
  d.Set("power", power)

  // Do network interfaces
  log.Printf("virtual_networks: %q\n", virtual_networks)
  nics := make([]map[string]interface{}, 0, 1)

  for nic := 0; nic < 3; nic++ {
    if virtual_networks[nic][0] != "" {
      out := make(map[string]interface{})
      out["virtual_network"] = virtual_networks[nic][0]
      out["mac_address"]     = virtual_networks[nic][1]
      out["nic_type"]        = virtual_networks[nic][2]
      nics = append(nics, out)
    }
  }

  d.Set("network_interfaces", nics)

  return nil
}
