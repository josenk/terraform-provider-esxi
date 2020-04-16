package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePortGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourcePortGroupCreate,
		Read:   resourcePortGroupRead,
		Delete: resourcePortGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVirtualSwitchImport,
		},
		Schema: map[string]*schema.Schema{
			"port_group_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the port group.",
				ForceNew:    true,
			},
			"virtual_switch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the virtual switch to bind this port group to.",
				ForceNew:    true,
			},
		},
	}
}
