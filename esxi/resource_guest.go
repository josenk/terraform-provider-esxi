package esxi

import (
  "fmt"
  "errors"
  "github.com/hashicorp/terraform/helper/schema"
  "strconv"
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
      //"esxi_virtual_network": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("esxi_virtual_network", nil),
      //    Description: "esxi virtual network.",
      //},
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
      //"guest_disk_type": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("guest_disk_type", nil),
      //    Description: "Guest guest disk type .",
      //},
      //"guest_storage": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("guest_storage", nil),
      //    Description: "Guest guest additional storage.",
      //},
      //"guest_nic_type": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("guest_nic_type", nil),
      //    Description: "Guest guest nic type.",
      //},
      //"guest_mac_address": &schema.Schema{
      //    Type:     schema.TypeString,
      //    Required: true,
      //    DefaultFunc: schema.EnvDefaultFunc("guest_mac_address", nil),
      //    Description: "Guest guest mac address.",
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
    },
  }
}

func resourceGUESTCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  clone_from_vm      := d.Get("clone_from_vm").(string)
  ovf_source         := d.Get("ovf_source").(string)
  disk_store         := d.Get("disk_store").(string)
  //esxi_virtual_network := d.Get("esxi_virtual_network").(string)
  resource_pool_name := d.Get("resource_pool_name").(string)
  guest_name         := d.Get("guest_name").(string)
  //guest_disk_type    := d.Get("guest_disk_type").(string)
  //guest_storage      := d.Get("guest_storage").(string)
  //guest_nic_type     := d.Get("guest_nic_type").(string)
  //guest_mac_address  := d.Get("guest_mac_address").(string)
  memsize            := d.Get("memsize").(string)
  numvcpus           := d.Get("numvcpus").(string)

  // Validations
  var src_path string

  if resource_pool_name == "ha-root-pool" {
    resource_pool_name = "/"
  }
  if string(resource_pool_name[0]) != "/" {
    resource_pool_name = "/" + resource_pool_name
  }

  if clone_from_vm != "" {
    src_path = fmt.Sprintf("vi://%s:%s@%s/%s", c.Esxi_username, c.Esxi_password, c.Esxi_hostname, clone_from_vm)
    fmt.Println("[Terraform-provider-esxi]   ")
  } else if ovf_source != "" {
    src_path = ovf_source
  } else {
    fmt.Println("[provider-esxi] Error: You must specify clone_from_vm or src_path as a source.")
    return errors.New("Error: You must specify clone_from_vm or src_path as a source.")
  }

  if _, err := strconv.Atoi(memsize); err != nil && memsize != "" {
    return errors.New("Error: memsize must be an integer")
  }
  if _, err := strconv.Atoi(numvcpus); err != nil && numvcpus != "" {
    return errors.New("Error: numvcpus must be an integer")
  }

  vmid, err := guestCREATE(c, guest_name, disk_store, src_path, resource_pool_name, memsize, numvcpus)
  if err == nil {
    d.SetId(vmid)
  } else {
    fmt.Println("Error: Unable to create guest.")
    return errors.New(vmid)
  }
  return nil
}

func resourceGUESTRead(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)

  guest_name, disk_store, resource_pool_name, memsize, numvcpus, err := guestREAD(c, d.Id())
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

  return nil
}

func resourceGUESTUpdate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
  memsize      := d.Get("memsize").(string)
  numvcpus     := d.Get("numvcpus").(string)

  err := guestUPDATE(c, d.Id(), memsize, numvcpus)

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
