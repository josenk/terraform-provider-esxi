package esxi

import (
  "fmt"
  "log"
  "github.com/hashicorp/terraform/helper/schema"
  "strings"
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
          ForceNew: true,
          DefaultFunc: schema.EnvDefaultFunc("resource_pool_name", nil),
          Description: "Resource Pool Name",
      },
      "cpu_min": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("cpu_min", nil),
          Description: "CPU minimum (in MHz).",
      },
      "cpu_min_expandable": &schema.Schema{
          Type:     schema.TypeBool,
          Required: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("cpu_min_expandable", "true"),
          Description: "Can pool borrow CPU resources from parent?",
      },
      "cpu_max": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("cpu_max", nil),
          Description: "CPU maximum (in MHz).",
      },
      "cpu_shares": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("cpu_shares", "normal"),
          Description: "CPU shares (low/normal/high/<custom>).",
      },
      "mem_min": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("mem_min", nil),
          Description: "Memory minimum (in MB).",
      },
      "mem_min_expandable": &schema.Schema{
          Type:     schema.TypeBool,
          Required: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("mem_min_expandable", "true"),
          Description: "Can pool borrow memory resources from parent?",
      },
      "mem_max": &schema.Schema{
          Type:     schema.TypeInt,
          Optional: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("mem_max", nil),
          Description: "Memory maximum (in MB).",
      },
      "mem_shares": &schema.Schema{
          Type:     schema.TypeString,
          Required: true,
          ForceNew: false,
          DefaultFunc: schema.EnvDefaultFunc("mem_shares", "normal"),
          Description: "Memory shares (low/normal/high/<custom>).",
      },
    },
  }
}

func resourceRESOURCEPOOLCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  var pool_id, parent_pool, resource_pool_name_cooked string

  resource_pool_name := d.Get("resource_pool_name").(string)
  cpu_min            := d.Get("cpu_min").(int)
  cpu_min_expandable := d.Get("cpu_min_expandable").(bool)
  cpu_max            := d.Get("cpu_max").(int)
  cpu_shares         := strings.ToLower(d.Get("cpu_shares").(string))
  mem_min            := d.Get("mem_min").(int)
  mem_min_expandable := d.Get("mem_min_expandable").(bool)
  mem_max            := d.Get("mem_max").(int)
  mem_shares         := strings.ToLower(d.Get("mem_shares").(string))

  if resource_pool_name == string('/') {
    return fmt.Errorf("Missing required resource_pool_name")
  }

  parent_pool               = "Resources"
  resource_pool_name_cooked = resource_pool_name

  if resource_pool_name[0] == '/' {
    return fmt.Errorf("Resource Pool Name cannot start with /")
  }
  i := strings.LastIndex(resource_pool_name, "/")
  if i > 2 {
    parent_pool = resource_pool_name[:i]
    resource_pool_name_cooked = resource_pool_name[i+1:]
  }

  pool_id,err := resourcePoolCREATE(c, resource_pool_name_cooked, cpu_min, cpu_min_expandable,
    cpu_max, cpu_shares, mem_min, mem_min_expandable, mem_max, mem_shares, parent_pool, )
  if err == nil {
    d.SetId(pool_id)
  } else {
    d.SetId("")
  }
  return nil
}

func resourceRESOURCEPOOLRead(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  cpu_min, cpu_min_expandable, cpu_max, cpu_shares, mem_min, mem_min_expandable, mem_max, mem_shares, resource_pool_name, err := resourcePoolREAD(c, d.Id())
  if err != nil {
    d.SetId("")
    return nil
  }

  d.Set("resource_pool_name", resource_pool_name)
  d.Set("cpu_min", cpu_min)
  d.Set("cpu_min_expandable", cpu_min_expandable)
  d.Set("cpu_max", cpu_max)
  d.Set("cpu_shares", cpu_shares)
  d.Set("mem_min", mem_min)
  d.Set("mem_min_expandable", mem_min_expandable)
  d.Set("mem_max", mem_max)
  d.Set("mem_shares", mem_shares)

  return nil
}

func resourceRESOURCEPOOLUpdate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  var pool_id string

  resource_pool_name := d.Get("resource_pool_name").(string)
  cpu_min            := d.Get("cpu_min").(int)
  cpu_min_expandable := d.Get("cpu_min_expandable").(bool)
  cpu_max            := d.Get("cpu_max").(int)
  cpu_shares         := strings.ToLower(d.Get("cpu_shares").(string))
  mem_min            := d.Get("mem_min").(int)
  mem_min_expandable := d.Get("mem_min_expandable").(bool)
  mem_max            := d.Get("mem_max").(int)
  mem_shares         := strings.ToLower(d.Get("mem_shares").(string))


  if resource_pool_name == string('/') {
    resource_pool_name = "Resources"
  }
  if resource_pool_name[0] == '/' {
    resource_pool_name = resource_pool_name[1:]
  }

  pool_id,err := resourcePoolUPDATE(c, d.Id(), cpu_min, cpu_min_expandable,
    cpu_max, cpu_shares, mem_min, mem_min_expandable, mem_max, mem_shares )
  if err == nil {
    log.Printf("[provider-esxi / resourceRESOURCEPOOLUpdate] success update: %s\n", pool_id)
  } else {
    d.SetId("")
    return nil
  }
  return nil
}

func resourceRESOURCEPOOLDelete(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  err := resourcePoolDELETE(c, d.Id())
  if err == nil {
    d.SetId("")
  }
  return nil
}
