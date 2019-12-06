provider "esxi" {
  esxi_hostname      = var.esxi_hostname
  esxi_hostport      = var.esxi_hostport
  esxi_username      = var.esxi_username
  esxi_password      = var.esxi_password
}

#
# Template for initial configuration bash script
#    template_file is a great way to pass variables to
#    cloud-init
data "template_file" "userdata_default" {
  template = file("userdata.tpl")
  vars = {
    HOSTNAME = var.vm_hostname
    HELLO    = "Hello EXSI World!"
  }
}

resource "esxi_guest" "vmtest" {
  guest_name         = var.vm_hostname
  disk_store         = var.disk_store

  network_interfaces {
     virtual_network = var.virtual_network
  }

  guestinfo = {
    "userdata.encoding" = "gzip+base64"
    "userdata"          = base64gzip(data.template_file.userdata_default.rendered)
  }

  #
  #  Specify an ovf file to use as a source.
  #
  ovf_source        = var.ovf_file

  ovf_property {
    key = "password"
    value = "Passw0rd1"
  }

  ovf_property {
    key = "hostname"
    value = "HelloWorld"
  }

  ovf_property {
    key = "user-data"
    value = base64encode(data.template_file.userdata_default.rendered)
  }
}
