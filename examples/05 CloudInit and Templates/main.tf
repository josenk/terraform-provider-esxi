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
#  cloud-init for vmware!
#  You must install it on your source VM before cloning it!
#    See https://github.com/akutz/cloud-init-vmware-guestinfo for more details.
#    and https://cloudinit.readthedocs.io/en/latest/topics/examples.html#
#
#
#    yum install https://github.com/akutz/cloud-init-vmware-guestinfo/releases/download/v1.1.0/cloud-init-vmware-guestinfo-1.1.0-1.el7.noarch.rpm
#    cloud-init clean
#########################################


#
# Template for initial configuration bash script
#    template_file is a great way to pass variables to
#    cloud-init
data "template_file" "Default" {
  template = "${file("userdata.tpl")}"
  vars = {
    HOSTNAME = "${var.vm_hostname}"
    HELLO    = "Hello World!"
  }
}


#########################################
#  ESXI Guest resource
#########################################
resource "esxi_guest" "Default" {
  guest_name         = "${var.vm_hostname}"
  disk_store         = "${var.disk_store}"

  clone_from_vm      = "Templates/centos7"

  network_interfaces = [
    {
      virtual_network = "${var.virtual_network}"
    },
  ]

  guestinfo = {
    userdata.encoding = "gzip+base64"
    userdata = "${base64gzip(data.template_file.Default.rendered)}"
  }

}
