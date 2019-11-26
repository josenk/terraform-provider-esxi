
output "ip" {
  value = esxi_guest.vmtest.ip_address
}

output "cloudinit" {
  value = data.template_file.userdata_default.rendered
}
