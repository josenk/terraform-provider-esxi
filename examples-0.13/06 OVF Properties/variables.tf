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

varaible "esxi_hostssl" {
  default = "443"
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
  default = "DiskStore01"
}

# Example downloaded from https://cloud-images.ubuntu.com/
variable "ovf_file" {
   default = "ubuntu-19.04-server-cloudimg-amd64.ova"
}

variable "vm_hostname" {
  default = "vmtest06"
}
