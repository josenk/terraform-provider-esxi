package esxi

import (
  "github.com/hashicorp/terraform/helper/schema"
)

func resourceRESOURCEPOOL() *schema.Resource {
  return &schema.Resource{
    Create: resourceRESOURCEPOOLCreate,
    Read:   resourceRESOURCEPOOLRead,
    Update: resourceRESOURCEPOOLUpdate,
    Delete: resourceRESOURCEPOOLDelete,
    Schema: map[string]*schema.Schema{
      "resource_pool_name": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: false,
          Description: "Resource Pool Name",
      },
      "cpu_min": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "CPU minimum (in MHz).",
      },
      "cpu_min_expandable": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Can pool borrow CPU resources from parent?",
      },
      "cpu_max": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "CPU maximum (in MHz).",
      },
      "cpu_shares": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "CPU shares (low/normal/high/<custom>).",
      },
      "mem_min": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Memory minimum (in MB).",
      },
      "mem_min_expandable": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Can pool borrow memory resources from parent?",
      },
      "mem_max": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Memory maximum (in MB).",
      },
      "mem_shares": &schema.Schema{
          Type:     schema.TypeString,
          Optional: true,
          ForceNew: false,
          Computed: true,
          Description: "Memory shares (low/normal/high/<custom>).",
      },
    },
  }
}
