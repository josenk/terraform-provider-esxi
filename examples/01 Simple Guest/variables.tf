#
#  See https://www.terraform.io/intro/getting-started/variables.html for more details.
#

variable "esxi_hostname" { default = "esxi" }
variable "esxi_hostport" { default = "22" }
variable "esxi_username" { default = "root" }
variable "esxi_password" { } # Unspecified will prompt
