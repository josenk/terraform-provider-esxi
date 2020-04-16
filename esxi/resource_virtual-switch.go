package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitch() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualSwitchCreate,
		Read:   resourceVirtualSwitchRead,
		Delete: resourceVirtualSwitchDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVirtualSwitchImport,
		},
		Schema: map[string]*schema.Schema{
			"virtual_switch_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Computed:    false,
				Description: "Virtual Switch Name",
			},
			// // uplinks
			// "network_adapters": {
			// 	Type:        schema.TypeList,
			// 	Required:    true,
			// 	Description: "The list of network adapters to bind to this virtual switch.",
			// 	Elem:        &schema.Schema{Type: schema.TypeString},
			// },
		},
	}
}
