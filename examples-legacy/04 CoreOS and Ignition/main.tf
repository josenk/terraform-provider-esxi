#########################################
#  ESXI Provider host/login details
#########################################
#
#   Use of variables here to hide/move the variables to a separate file
#
provider "esxi" {
  version       = "~> 1.3"
  esxi_hostname = "${var.esxi_hostname}"
  esxi_hostport = "${var.esxi_hostport}"
  esxi_username = "${var.esxi_username}"
  esxi_password = "${var.esxi_password}"
}


#########################################
#  coreos and Ignition
#    See: https://coreos.com/ignition/docs/latest/
#
#    - To use this example, you must download the latest ova from coreos site.
#    - curl -LO https://stable.release.core-os.net/amd64-usr/current/coreos_production_vmware_ova.ova
#########################################


data "ignition_user" "cluster_user" {
  name = "core"
  # used to calculate a password hash below if you want password: mkpasswd --method=SHA-512 --rounds=4096
  #password_hash = "$6$rounds=4096$v28up9xdAO$aLCyf/FfGU72QsOlJN60CyXH5yuyJ/f6WWeW3wyvPjHt4uDOUFbFCxchrf9FCUkUdng7bwZbonBk7aOFQ8Bcm0" # test1234
  ssh_authorized_keys = ["${file("~/.ssh/id_rsa.pub")}"]
}

data "ignition_systemd_unit" "example" {
  name = "example.service"
  content = "${file("example.service")}"
}

data "ignition_config" "coreos" {
  users = [
    "${data.ignition_user.cluster_user.id}"
  ]

  systemd = [
    "${data.ignition_systemd_unit.example.id}"
  ]
}

resource "esxi_guest" "coreos" {
  guest_name = "${terraform.workspace}-coreos"
  disk_store = "${var.disk_store}"

  guestinfo = {
    coreos.config.data.encoding = "base64"
    coreos.config.data = "${base64encode(data.ignition_config.coreos.rendered)}"
  }

  ovf_source = "coreos_production_vmware_ova.ova"
  power = "on"
  memsize  = "2048"
  numvcpus = "2"
  network_interfaces = [
    {
      virtual_network = "VM Network" # Required for each network interface, Specify the Virtual Network name.
    },
  ]

  connection {
    host     = "${self.ip_address}"
    type     = "ssh"
    user     = "core"
    private_key = "${file("~/.ssh/id_rsa")}"
    timeout = "60s"
  }

  provisioner "remote-exec" {
    inline = [
      "echo success",
      #"update_engine_client -update",
    ]
  }
}
