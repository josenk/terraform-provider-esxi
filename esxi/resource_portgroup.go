package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourcePORTGROUP() *schema.Resource {
	return &schema.Resource{
		Create: resourcePORTGROUPCreate,
		Read:   resourcePORTGROUPRead,
		Update: resourcePORTGROUPUpdate,
		Delete: resourcePORTGROUPDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePORTGROUPImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "portgroup name.",
			},
			"vswitch": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "vswitch name.",
			},
			"vlan": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Computed:     true,
				Description:  "portgroup vlan.",
				ValidateFunc: validation.IntBetween(0, 4095),
			},
		},
	}
}
