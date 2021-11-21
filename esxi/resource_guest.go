package esxi

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceGUEST() *schema.Resource {
	return &schema.Resource{
		Create: resourceGUESTCreate,
		Read:   resourceGUESTRead,
		Update: resourceGUESTUpdate,
		Delete: resourceGUESTDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGUESTImport,
		},
		Schema: map[string]*schema.Schema{
			"clone_from_vm": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "Source vm path on esxi host to clone.",
			},
			"host_ovf": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "Path on esxi host of ovf files.",
			},
			"ovf_source": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "Local path to source ovf files.",
			},
			"disk_store": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "esxi diskstore for boot disk.",
			},
			"resource_pool_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "/",
				Description: "Resource pool name to place guest.",
			},
			"guest_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "esxi guest name.",
			},
			"boot_disk_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "thin",
				Description: "Guest boot disk type. thin, zeroedthick, eagerzeroedthick",
			},
			"boot_disk_size": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Default:     nil,
				Description: "Guest boot disk size. Will expand boot disk to this size.",
			},
			"memsize": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest guest memory size.",
			},
			"numvcpus": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest guest number of virtual cpus.",
			},
			"virthwver": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest Virtual HW version.",
			},
			"guestos": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest OS type.",
			},
			"network_interfaces": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"virtual_network": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							Computed: true,
						},
						"mac_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							Computed: true,
						},
						"nic_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							Computed: true,
						},
						"ovf_network": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							Computed: true,
						},
					},
				},
			},
			"power": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest power state.",
			},
			//  Calculated only, you cannot overwrite this.
			"ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP address reported by VMware tools.",
			},
			"guest_startup_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The amount of guest uptime, in seconds, to wait for an available IP address on this virtual machine.",
				ValidateFunc: validation.IntBetween(0, 600),
			},
			"guest_shutdown_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The amount of time, in seconds, to wait for a graceful shutdown before doing a forced power off.",
				ValidateFunc: validation.IntBetween(0, 600),
			},
			"virtual_disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: false,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"virtual_disk_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"slot": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "SCSI_Ctrl:SCSI_id.    Range  '0:1' to '0:15'.   SCSI_id 7 is not allowed.",
						},
					},
				},
			},
			"ovf_properties": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: false,
				ForceNew: true,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							Computed: false,
						},
					},
				},
			},
			"ovf_properties_timer": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The amount of time, in seconds, to wait for the guest to boot and run ovf_properties.",
				ValidateFunc: validation.IntBetween(0, 6000),
			},
			"notes": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Guest notes (annotation).",
			},
			"guestinfo": &schema.Schema{
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "pass data to VM",
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceGUESTCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	log.Printf("[resourceGUESTCreate]\n")

	var virtual_networks [10][4]string
	var virtual_disks [60][2]string
	var src_path string
	var tmpint, i, virtualDiskCount, ovfPropsCount, guest_shutdown_timeout, ovf_properties_timer int
	var ovf_properties map[string]string

	clone_from_vm := d.Get("clone_from_vm").(string)
	ovf_source := d.Get("ovf_source").(string)
	disk_store := d.Get("disk_store").(string)
	resource_pool_name := d.Get("resource_pool_name").(string)
	guest_name := d.Get("guest_name").(string)
	boot_disk_type := d.Get("boot_disk_type").(string)
	boot_disk_size := d.Get("boot_disk_size").(string)
	memsize := d.Get("memsize").(string)
	numvcpus := d.Get("numvcpus").(string)
	virthwver := d.Get("virthwver").(string)
	guestos := d.Get("guestos").(string)
	notes := d.Get("notes").(string)
	power := d.Get("power").(string)

	if d.Get("guest_startup_timeout").(int) > 0 {
		d.Set("guest_startup_timeout", d.Get("guest_startup_timeout").(int))
	} else {
		d.Set("guest_startup_timeout", 120)
	}
	if d.Get("guest_shutdown_timeout").(int) > 0 {
		d.Set("guest_shutdown_timeout", d.Get("guest_shutdown_timeout").(int))
		guest_shutdown_timeout = d.Get("guest_shutdown_timeout").(int)
	} else {
		d.Set("guest_shutdown_timeout", 20)
	}
	if d.Get("ovf_properties_timer").(int) > 0 {
		d.Set("ovf_properties_timer", d.Get("ovf_properties_timer").(int))
		ovf_properties_timer = d.Get("ovf_properties_timer").(int)
	} else {
		d.Set("ovf_properties_timer", 90)
		ovf_properties_timer = 90
	}

	guestinfo, ok := d.Get("guestinfo").(map[string]interface{})
	if !ok {
		return errors.New("guestinfo is wrong type")
	}

	// Validations
	if resource_pool_name == "ha-root-pool" {
		resource_pool_name = "/"
	}

	if clone_from_vm != "" {
		password := url.QueryEscape(c.esxiPassword)
		src_path = fmt.Sprintf("vi://%s:%s@%s:%s/%s", c.esxiUserName, password, c.esxiHostName, c.esxiHostSSLport, clone_from_vm)
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
	if boot_disk_type != "thin" && boot_disk_type != "zeroedthick" && boot_disk_type != "eagerzeroedthick" {
		return errors.New("Error: boot_disk_type must be thin, zeroedthick or eagerzeroedthick")
	}

	//  Validate boot_disk_size.
	if _, err := strconv.Atoi(boot_disk_size); err != nil && boot_disk_size != "" {
		return errors.New("Error: boot_disk_size must be an integer")
	}
	tmpint, _ = strconv.Atoi(boot_disk_size)
	if (tmpint < 1 || tmpint > 62000) && boot_disk_size != "" {
		return errors.New("Error: boot_disk_size must be an > 1 and < 62000")
	}

	//  Validate lan adapters
	lanAdaptersCount := d.Get("network_interfaces.#").(int)
	if lanAdaptersCount > 10 {
		lanAdaptersCount = 10
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

		if attr, ok := d.Get(prefix + "ovf_network").(string); ok && attr != "" {
			virtual_networks[i][3] = d.Get(prefix + "ovf_network").(string)
		}
	}

	//  Validate virtual_disks
	virtualDiskCount, ok = d.Get("virtual_disks.#").(int)
	if !ok {
		virtualDiskCount = 0
		virtual_disks[0][0] = ""
	}

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
			validateVirtualDiskSlot(virtual_disks[i][1])
			result := validateVirtualDiskSlot(virtual_disks[i][1])
			if result != "ok" {
				return errors.New(result)
			}
		}
	}

	//  Parse ovf properties, if any
	ovfPropsCount, ok = d.Get("ovf_properties.#").(int)
	if !ok {
		ovfPropsCount = 0
	} else {
		ovf_properties = make(map[string]string)
	}

	for i = 0; i < ovfPropsCount; i++ {
		prefix := fmt.Sprintf("ovf_properties.%d.", i)

		if key, ok := d.Get(prefix + "key").(string); ok && key != "" {

			if value, ok := d.Get(prefix + "value").(string); ok && value != "" {
				ovf_properties[key] = value
			}
		}
	}

	vmid, err := guestCREATE(c, guest_name, disk_store, src_path, resource_pool_name, memsize,
		numvcpus, virthwver, guestos, boot_disk_type, boot_disk_size, virtual_networks,
		virtual_disks, guest_shutdown_timeout, ovf_properties_timer, notes, guestinfo, ovf_properties)
	if err != nil {
		tmpint, _ = strconv.Atoi(vmid)
		if tmpint > 0 {
			d.SetId(vmid)
		}
		return err
	}

	//  set vmid
	d.SetId(vmid)

	if power == "on" || power == "" {
		_, err = guestPowerOn(c, vmid)
		if err != nil {
			return errors.New("Failed to power on.")
		}
	}
	d.Set("power", "on")

	// Refresh
	return resourceGUESTRead(d, m)
}
