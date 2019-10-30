#
#  See https://www.terraform.io/intro/getting-started/variables.html for more details.
#

#  Change these defaults to fit your needs!

variable "esxi_hostname" {
  default = "esxi"
}

variable "esxi_hostport" {
  default = "22"
}

variable "esxi_username" {
  default = "root"
}

variable "esxi_password" {
  # Unspecified will prompt
}

variable "virtual_network" {
  default = "VM Network"
}

variable "disk_store" {
  default = "ESXI"
}

# Application specific variables

variable "vm_clone_from" {
  default = "templateU18D"
}

variable "vm_ovf_local_path" {
   default = "/home/slavko/personal/ESXI/bionic/bionic.ova"
}

variable "vm_hostname" {
  default = "vmtest"
}
