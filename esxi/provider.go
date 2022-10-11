package esxi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"esxi_hostname": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("", "esxi"),
				Description: "The esxi hostname or IP address.",
			},
			"esxi_hostport": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("esxi_hostport", "22"),
				Description: "ssh port.",
			},
			"esxi_hostssl": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("esxi_hostssl", "443"),
				Description: "ssl port.",
			},
			"esxi_username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("esxi_username", "root"),
				Description: "esxi ssh username.",
			},
			"esxi_password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("esxi_password", "unset"),
				Description: "esxi ssh password.",
			},
			"esxi_remote_ovftool_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ovftool path on ESXi host",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"esxi_guest":         resourceGUEST(),
			"esxi_resource_pool": resourceRESOURCEPOOL(),
			"esxi_virtual_disk":  resourceVIRTUALDISK(),
			"esxi_vswitch":       resourceVSWITCH(),
			"esxi_portgroup":     resourcePORTGROUP(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		esxiHostName:          d.Get("esxi_hostname").(string),
		esxiHostSSHport:       d.Get("esxi_hostport").(string),
		esxiHostSSLport:       d.Get("esxi_hostssl").(string),
		esxiUserName:          d.Get("esxi_username").(string),
		esxiPassword:          d.Get("esxi_password").(string),
		esxiRemoteOvfToolPath: d.Get("esxi_remote_ovftool_path").(string),
	}

	if err := config.validateEsxiCreds(); err != nil {
		return nil, err
	}

	return &config, nil
}
