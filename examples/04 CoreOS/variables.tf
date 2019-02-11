#
#  See https://www.terraform.io/intro/getting-started/variables.html for more details.
#

variable "esxi_hostname" {
  default = "10.5.5.5"
}

variable "esxi_hostport" {
  default = "22"
}

variable "esxi_username" {
  default = "root"
}

variable "esxi_password" {
  #default = ""
} # Unspecified will prompt
