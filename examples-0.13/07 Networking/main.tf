#########################################
#  ESXI Provider host/login details
#########################################
#
#   Use of variables here to hide/move the variables to a separate file
#
provider "esxi" {
  esxi_hostname = var.esxi_hostname
  esxi_hostport = var.esxi_hostport
  esxi_hostssl  = var.esxi_hostssl
  esxi_username = var.esxi_username
  esxi_password = var.esxi_password
}

#########################################
#  ESXI vSwitch resource
#########################################
#
#  Example vswitch with defaults.
#  Uncommend the uplink block to connect this vswitch to your nic.
#
resource "esxi_vswitch" "myvswitch" {
  name = "My vSwitch"
  #uplink {
  #  name = "vmnic1"
  #}
}

#########################################
#  ESXI Port Group resource
#########################################
#
#  Example port group with default, connecting to the above vswitch.
#
resource "esxi_portgroup" "myportgroup" {
  name = "My Port Group"
  vswitch = esxi_vswitch.myvswitch.name
}

#########################################
#  ESXI Guest resource
#########################################
#
#  This Guest VM is "bare-metal".   It will be powered on by default
#  by terraform, but it will not boot to any OS.   It will however attempt
#  to network boot on the port group configured above.
#
resource "esxi_guest" "vmtest01" {
  guest_name = "vmtest01"
  disk_store = "DS_001"
  network_interfaces {
    virtual_network = esxi_portgroup.myportgroup.name  # Connecting to the above portgroup
  }
}
