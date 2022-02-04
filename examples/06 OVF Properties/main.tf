provider "esxi" {
  esxi_hostname      = var.esxi_hostname
  esxi_hostport      = var.esxi_hostport
  esxi_hostssl       = var.esxi_hostssl
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
    HELLO    = "Hello ESXi World!"
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

  #
  #  Specify ovf_properties specific to the source ovf/ova.
  #    Use ovftool <filename>.ova to get details of which ovf_properties are available.
  #
  ovf_properties {
    key = "hostname"
    value = "firstboot"
  }

  ovf_properties {
    key = "user-data"
    value = base64encode(data.template_file.userdata_default.rendered)
  }

  #
  #  Default ovf_properties_timer is 90 seconds.  ovf_properties are injected on first
  #  boot.  This value should be high enough to allow the vmguest to fully boot to 
  #  a linux prompt.  The second boot is needed to configure the vmguest as
  #  specified. (cpus, memory, adding or expanding disks, etc...)
  # 
  #ovf_properties_timer = 90
}
