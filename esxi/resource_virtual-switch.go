package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVirtualSwitch() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualSwitchCreate,
		Read:   resourceVirtualSwitchRead,
		Update: resourceVirtualSwitchUpdate,
		Delete: resourceVirtualSwitchDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVirtualSwitchImport,
		},
		Schema: map[string]*schema.Schema{
			"virtual_switch_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Computed:    false,
				Description: "Virtual Switch Name",
			},
		},
	}
}
