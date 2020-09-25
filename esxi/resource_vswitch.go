package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceVSWITCH() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSWITCHCreate,
		Read:   resourceVSWITCHRead,
		Update: resourceVSWITCHUpdate,
		Delete: resourceVSWITCHDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSWITCHImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "vswitch name.",
			},
			"ports": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				Description:  "vswitch number of ports.",
				ValidateFunc: validation.IntBetween(1, 4096),
			},
			"mtu": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Computed:     true,
				Description:  "vswitch mtu.",
				ValidateFunc: validation.IntBetween(1280, 9000),
			},
			"link_discovery_mode": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "vswitch Link Discovery Mode.",
			},
			"promiscuous_mode": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Promiscuous mode (true=Accept/false=Reject).",
			},
			"mac_changes": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "MAC address changes (true=Accept/false=Reject).",
			},
			"forged_transmits": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Forged transmits (true=Accept/false=Reject).",
			},
			"uplink": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}
