provider "esxi" {
  esxi_hostname      = var.esxi_hostname
  esxi_hostport      = var.esxi_hostport
  esxi_hostssl       = var.esxi_hostssl
  esxi_username      = var.esxi_username
  esxi_password      = var.esxi_password

  esxi_remote_ovftool_path = "/vmfs/volumes/datastore1/ovftool/ovftool"
}

resource "esxi_guest" "vmtest" {
  guest_name         = var.vm_hostname
  disk_store         = var.disk_store

  network_interfaces {
     virtual_network = var.virtual_network
  }

  #
  #  Specify an ovf file to use as a source.
  #
  ovf_source        = "host_ovf:///vmfs/volumes/datastore1/ovas/ubuntu-22.04-server-cloudimg-amd64.ova"
}
