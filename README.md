Terraform-provider-esxi plugin
==============================
This is a Terraform plugin that adds a VMware ESXi provider support.  This allows Terraform to control and provision VMs directly on an ESXi hypervisor without a need for vCenter or VShpere.   ESXi hypervisor is a free download from VMware!
>https://www.vmware.com/go/get-free-esxi

What's New:
-----------
* Resource pools should be fully functional.  Terraform will Add, Delete, Update & Read Resource Pools.
* Guest vms framework is there.  Terraform will Add, Delete & Read the Guest vms.


Documentation:
-------------
* This is an early development release!!!   There is basic functionality and some validation.   Error messages may be limited.  I'll add features and update documentation as time permits...

* If you don't know terraform, I highly recommend you read through the introduction on the hashicorp website.
>https://www.terraform.io/intro/getting-started/install.html

* VMware Configuration Maximums tool.
>https://configmax.vmware.com/guest


Features and Compatibility
--------------------------
* Source image can be a clone of a VM or local vmx, ovf, ova file. This provider uses ovftool, so there should be a wide compatibility.
* Supports adding your VM to Resource Pools to partition CPU and memory usage from other VMs on your ESXi host.

Requirements
------------
1. This is a Terraform plugin, so you need Terraform installed...  :-)
2. This plugin requires ovftool from VMware.  Download from VMware website.  NOTE: ovftool installer for windows doesn't put ovftool.exe in your path.  You can manually set your path, or install ovftool in the \HashiCorp\Vagrant\bin directory.
>https://www.vmware.com/support/developer/ovf/
3. You MUST enable ssh access on your ESXi hypervisor.
  * Google 'How to enable ssh access on esxi'
4. In general, you should know how to use terraform, esxi and some networking...

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
2. `vi main.tf`  # Use the contents of this example main.tf as a template. Specify provider parameters to access your ESXi host.  Modify the resources for resource pools and guest vm.

```
provider "esxi" {
  esxi_hostname      = "esxi"
  esxi_hostport      = "22"
  esxi_username      = "root"
  esxi_password      = "MyPassword"
}
resource "esxi_resource_pool" "MyPool" {
  resource_pool_name = "MyPool"
  cpu_min            = "100"
  mem_min            = "200"
}
resource "esxi_guest" "vmtest" {
  depends_on         = ["esxi_resource_pool_name.MyPool"]
  guest_name         = "v-test"

  #
  #  Specify an existing guest to clone, an ovf source, or neither to build a guest vm from scratch
  #
  #clone_from_vm      = "Templates/centos7"
  #ovf_source        = "/my_local_system_path/centos-7-min/centos-7.vmx"

  disk_store         = "MyDiskStore"
  resource_pool_name = "MyPool"
  network_interfaces = [
    {
      virtual_network = "VM Network"
      mac_address     = "00:50:56:a1:b1:c1"
      nic_type        = "e1000"
    },
    {
      virtual_network = "VM Network 2"
      nic_type        = "e1000"
    },
  ]
}
```

Basic usage
-----------
3. `terraform init`
4. `terraform plan`
5. `terraform apply`
6. `terraform show`
7. `terraform destroy`

Configuration reference
-----------------------
* provider "esxi"
  * esxi_hostname - Required
  * esxi_hostport - Optional - Default "22"
  * esxi_username - Optional - Default "root"
  * esxi_password - Required


* resource "esxi_resource_pool"
  * resource_pool_name - Required - The Resource Pool name.
  * cpu_min - Optional
  * cpu_min_expandable - Optional
  * cpu_max - Optional           
  * cpu_shares - Optional        
  * mem_min - Optional          
  * mem_min_expandable - Optional
  * mem_max - Optional           
  * mem_shares - Optional


* resource "esxi_guest"
  * guest_name - Required - The Guest name.
  * boot_disk_type - Optional - Guest boot disk type. Default 'thin'.  Available thin, thick, eagerzeroedthick.
  * boot_disk_size - Optional - Boot disk size.   If cloning a vm or using ovf_source, then this will grow the boot disk to this size.
  * clone_from_vm - Source vm to clone. Mutually exclusive with ovf_source option.     
  * ovf_source - ovf files to use as a source. Mutually exclusive with clone_from_vm option.      
  * disk_store - Required - Esxi Disk Store where guest vm will be created.    
  * resource_pool_name - Optional - Any resource pool name.     
  * memsize - Optional - Memory size in MB.  (ie, 1024 == 1GB). See esxi documentation for limits.
  * numvcpus - Optional - Number of virtual cpus.  See esxi documentation for limits.
  * virthwver - Optional - esxi guest virtual HW version.  See esxi documentation for compatible values.
  * network_interfaces - Array of network interfaces.
    * virtual_network - Required - esxi virtual network name configured on esxi host.
    * mac_address - Optional -  If not set, mac_address will be generated by esxi.
    * nic_type - Optional -  default 'e1000'.  See esxi documentation for compatibility list.
  * power - Optional - on/true, off/false.    

Known issues with vmware_esxi
-----------------------------
* More features coming.


Version History
---------------
* 0.0.8 Add virthwver.
* 0.0.7 build vmx from scratch if no source is specified
* 0.0.6 Add power resource.
* 0.0.5 Add network_interfaces resource.
* 0.0.4 Add more stuff.
* 0.0.3 Add memory and numvcpus resource.
      Add support to update some guests params.
* 0.0.2 Add Resource Pool resource.
* 0.0.1 Init release
