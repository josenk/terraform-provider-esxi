#########################################
#  ESXI Provider host/login details
#########################################
#
#   Use of variables here to hide/move the variables to a separate file
#
provider "esxi" {
  esxi_hostname  = "${var.esxi_hostname}"
  esxi_hostport  = "${var.esxi_hostport}"
  esxi_username  = "${var.esxi_username}"
  esxi_password  = "${var.esxi_password}"
}


#########################################
#  ESXI Guest resource
#########################################
#
#  This Guest VM is a clone of an existing Guest VM named "centos7" (must exist and
#  be powered off), located in the "Templates" resource pool.  vmtest02 will be powered
#  on by default by terraform.  The virtual network "VM Network", must already exist on
#  your esxi host!
#
resource "esxi_guest" "vmtest02" {
  guest_name         = "vmtest02"
  disk_store         = "DS_001"
  guestos            = "centos-64"

  boot_disk_type     = "thin"
  boot_disk_size     = "35"

  memsize            = "2048"
  numvcpus           = "2"
  resource_pool_name = "/"
  power              = "on"

  #  clone_from_vm uses ovftool to clone an existing Guest on your esxi host.  This example will clone a Guest VM named "centos7", located in the "Templates" resource pool.
  #  ovf_source uses ovftool to produce a clone from an ovf or vmx image. (typically produced using the ovf_tool).
  #    Basically clone_from_vm clones from sources on the esxi host and ovf_source clones from sources on your local hard disk.
  #    These two options are mutually exclusive.
  clone_from_vm      = "Templates/centos7"
  #ovf_source        = "/my_local_system_path/centos-7-min/centos-7.vmx"

  network_interfaces = [
    {
      virtual_network = "VM Network"
      mac_address     = "00:50:56:a1:b1:c2"
      nic_type        = "e1000"

    },
  ]

  guest_startup_timeout  = 45
  guest_shutdown_timeout = 30
}
