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

variable "esxi_hostssl" {
  default = "443"
}

variable "esxi_username" {
  default = "root"
}

variable "esxi_password" { # Unspecified will prompt 
}

variable "vmIP" {
  default = "10.10.10.10/24"
}

variable "vmGateway" {
  default = "10.10.10.1"
}

variable "nameserver" {
  default = "8.8.8.8"
}