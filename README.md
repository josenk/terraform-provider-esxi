Terraform-provider-esxi plugin
==============================
This is a Terraform plugin that adds a VMware ESXi provider support.  This allows Terraform to control and provision VMs directly on an ESXi hypervisor without a need for vCenter or VShpere.   ESXi hypervisor is a free download from VMware!
>https://www.vmware.com/go/get-free-esxi

Documentation:
-------------
* This is an early development release!!!   There is only some basic functionality and very little validation.   Error messages are limited.  I'll add features and update documentation as time permits...

Features and Compatibility
--------------------------
* Source image can be a clone of a VM or local vmx, ovf, ova file. This provider uses ovftool, so there should be a wide compatibility.
* Supports adding your VM to Resource Pools to partition CPU and memory usage from other VMs on your ESXi host.
* Disks can be provisioned using thin, thick or eagerzeroedthick.

Requirements
------------
1. This is a Terraform plugin, so you need Terraform installed...  :-)
2. This plugin requires ovftool from VMware.  Download from VMware website.  NOTE: ovftool installer for windows doesn't put ovftool.exe in your path.  You can manually set your path, or install ovftool in the \HashiCorp\Vagrant\bin directory.
>https://www.vmware.com/support/developer/ovf/
3. You MUST enable ssh access on your ESXi hypervisor.
  * Google 'How to enable ssh access on esxi'
4. In general, you should know how to use vagrant, esxi and some networking...

Why this plugin?
----------------
Not everyone has vCenter, vSphere, expensive APIs...  These cost $$$.  ESXi is free!

How to install
--------------
Download and install Terraform on your local system using instructions from https://www.terraform.io/downloads.html.
Download this plugin from github and place a copy of it in SOMEDIR.

How to use and configure a main.tf file
---------------------------------------

1. cd SOMEDIR
2. `vi main.tf`  # Replace the contents of main.tf with the following example. Specify parameters to access your ESXi host, guest and local preferences.

```
provider "esxi" {
  esxi_hostname  = "esxi"
  esxi_hostport  = "22"
  esxi_username  = "root"
  esxi_password  = "MyPassword"
}
resource "esxi_guest" "vmtest" {
  guest_name = "v-test"
  esxi_disk_store = "MyDiskStore"
  esxi_resource_pool = "Terraform"

  # Use clone_from_vm or ovf_source as a source.
  clone_from_vm = "Templates/centos7"
  #ovf_source = "/u1/devel/terraform/centos-7-min/centos-7.vmx"
}
```

Basic usage
-----------
3. `terraform init`
4. `terraform plan`
5. `terraform apply`
6. `terraform show`
7. `terraform destroy`

Known issues with vmware_esxi
-----------------------------
* Limited Features 
* Passwords are stored in clear-text in main.tf.
* Sources (ovf_source) must have no networks configured.
* Guests are not powered on

Version History
---------------
0.0.1 Init release
