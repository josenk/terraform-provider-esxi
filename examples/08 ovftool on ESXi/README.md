# Terraform esxi Provider (08 ovftool on ESXi)
---

Since we have SSH enabled on ESXi, it's possible to install ovftool directly on the ESXi host and avoid large network copies
by hosting the OVA/OVFs locally. 

To use this functionality, perform the following steps:
1. Install ovftool somewhere on the ESXi host. For example, ``/vmfs/volumes/datastore1/ovftool`` (directory)
2. Add the ``esxi_remote_ovftool_path`` to provider config:
```
provider "esxi" {
  esxi_hostname      = "${var.esxi_server}"
  esxi_username      = "${var.esxi_user}"
  esxi_password      = "${var.esxi_password}"
  esxi_remote_ovftool_path = "/vmfs/volumes/datastore1/ovftool/ovftool"
}
```
3. Install an OVA somewhere on the ESXi host. For example, /vmfs/volumes/datastore1/ovas/ubuntu-22.04-server-cloudimg-amd64.ova
4. Use the ``host_ovf://`` prefix to tell the plugin where to find the local image. Example:
```
resource "esxi_guest" "vmtest" {
  ...
   ovf_source = "host_ovf:///vmfs/volumes/datastore1/isos/ubuntu-22.04-server-cloudimg-amd64.ova"
}
```

Now, when creating the vmtest instance, terraform will run the ovftool on the ESXi host directly and will pull the image on that
same host, avoiding massive copies over the network. 

Note that ovf_properties and guestinfo directives will work as expected. We highly recommend using guestinfo directive w/ cloud-init
configs where possible (anything using cloud-init 21.2+ should work) and avoid the overhead of rebooting the instance multiple
times.