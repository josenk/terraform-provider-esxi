Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)


Requirements
------------
-   [Terraform](https://www.terraform.io/downloads.html) 0.11.x+
-   [Go](https://golang.org/doc/install) 1.11+ (to build the provider plugin)
-   [ovftool](https://www.vmware.com/support/developer/ovf/) from VMware.  NOTE: ovftool installer for windows doesn't put ovftool.exe in your path.  You will need to manually set your path.
-   You MUST enable ssh access on your ESXi hypervisor.
  * Google 'How to enable ssh access on esxi'
-   In general, you should know how to use terraform, esxi and some networking...
  * You will most likely need a DHCP server on your primary network if you are deploying VMs with public OVF/OVA/VMX images.  (Sources that have unconfigured primary interfaces.)
- The source OVF/OVA/VMX images must have open-vm-tools or vmware-tools installed to properly import an IPaddress.  (you need this to run provisioners)


Building The Provider
---------------------
In general, you don't normally need to build the provider unless you are planning to make changes to it.  A release can be downloaded from github, or can automatically be downloaded with terraform 0.13+.

You first must set your GOPATH.   If you are unsure, please review the documentation at.
>https://github.com/golang/go/wiki/SettingGOPATH


Clone repository to: `$GOPATH/src/github.com/josenk/terraform-provider-esxi`

```sh

mkdir $HOME/go
export GOPATH="$HOME/go"

go get -u -v golang.org/x/crypto/ssh
go get -u -v github.com/hashicorp/terraform
go get -u -v github.com/josenk/terraform-provider-esxi

cd $GOPATH/src/github.com/josenk/terraform-provider-esxi
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -extldflags "-static"' -o terraform-provider-esxi_`cat version`

sudo cp terraform-provider-esxi_`cat version` /usr/local/bin
```


Terraform-provider-esxi plugin
==============================
* This is a Terraform plugin that adds a VMware ESXi provider support.  This allows Terraform to control and provision VMs directly on an ESXi hypervisor without a need for vCenter or VShpere.   ESXi hypervisor is a free download from VMware!
>https://www.vmware.com/go/get-free-esxi

* If you don't know terraform, I highly recommend you read through the introduction on the hashicorp website.
>https://www.terraform.io/intro/getting-started/install.html

* VMware Configuration Maximums tool.
>https://configmax.vmware.com/guest


What's New:
-----------
* v1.9.0 Changed default hwversion from 8 to 13.  NOTE that this is a possible breaking change if you are using an old ESXi version.
* v1.8.0 added vswitch and portgroup resources.
* v1.7.1 added Terraform 0.13 support.  This provider is now in the terraform registry.
>https://registry.terraform.io/providers/josenk/esxi


Features and Compatibility
--------------------------
* Source image can be a clone of a VM or local vmx, ovf, ova file. This provider uses ovftool, so there should be a wide compatibility.
* Supports adding your VM to Resource Pools to partition CPU and memory usage from other VMs on your ESXi host.
* Terraform will Create, Destroy, Update & Import Resource Pools.
* Terraform will Create, Destroy, Update & Import Guest VMs.
* Terraform will Create, Destroy, Update & Import Extra Storage for Guests.
* Terraform will Create, Destroy, Update & Import vSwitches.
* Terraform will Create, Destroy, Update & Import Port Groups.


This is a provider!  NOT a provisioner.
---------------------------------------
* This plugin does not configure your guest VM, it creates it.
* To configure your guest VM after it's built, you need to use a provisioner.
  * Refer to Hashicorp list of provisioners: https://www.terraform.io/docs/provisioners/index.html
* To help you get started, there is are examples in a separate repo I created.   You can create a Pull Request if you would like to contribute.
  * https://github.com/josenk/terraform-provider-esxi-wiki


Vagrant vs Terraform.
---------------------
If you are using vagrant as a deployment tool (infa as code), you may want to consider a better tool.  Terraform.  Vagrant is better for development environments, while Terraform is better at managing infrastructure.  Please give my terraform plugin a try and give me some feedback.  What you're trying to do, what's missing, what works, what doesn't work, etc...
>https://www.vagrantup.com/intro/vs/terraform.html
>https://github.com/josenk/terraform-provider-esxi
>https://github.com/josenk/vagrant-vmware-esxi


Why this plugin?
----------------
Not everyone has vCenter, vSphere, expensive APIs...  These cost $$$.  ESXi is free!


How to install
--------------
* Install terraform
  * Download and install Terraform on your local system using instructions from https://www.terraform.io/downloads.html.
* Automatic install
  * Add the required_providers block to your terraform project.
```
terraform {
  required_version = ">= 0.13"
  required_providers {
    esxi = {
      source = "registry.terraform.io/josenk/esxi"
      #
      # For more information, see the provider source documentation:
      # https://github.com/josenk/terraform-provider-esxi
      # https://registry.terraform.io/providers/josenk/esxi
    }
  }
}
```
* Manual installation (Terraform 0.11.x or 0.12.x only)
  * Download pre-built binaries from https://github.com/josenk/terraform-provider-esxi/releases.  Place a copy of it in your path or current directory of your terraform project.



How to use and configure a main.tf file
---------------------------------------
1. cd SOMEDIR
2. `vi main.tf`  # Use the contents of this example main.tf as a template. Specify provider parameters to access your ESXi host.  Modify the resources for resource pools and guest vm.

```
terraform {
  required_version = ">= 0.12"
}

provider "esxi" {
  esxi_hostname      = "esxi"
  esxi_hostport      = "22"
  esxi_hostssl       = "443"
  esxi_username      = "root"
  esxi_password      = "MyPassword"
}

resource "esxi_guest" "vmtest" {
  guest_name         = "vmtest"
  disk_store         = "MyDiskStore"

  #
  #  Specify an existing guest to clone, an ovf source, or neither to build a bare-metal guest vm.
  #
  #clone_from_vm      = "Templates/centos7"
  #ovf_source        = "/local_path/centos-7.vmx"

  network_interfaces {
    virtual_network = "VM Network"
  }
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
  * esxi_hostport - Optional - Default "22".
  * esxi_hostssl - Optional - Default "443".
  * esxi_username - Optional - Default "root".
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


* resource "esxi_virtual_disk"
  * virtual_disk_disk_store - Required - esxi Disk Store where virtual disk will be created.
  * virtual_disk_dir - Required - A subdirectory to contain the virtual disk. (Can be the same as guest_name)
  * virtual_disk_name - Optional - Virtual Disk Name.  (ext must be .vmdk)
  * virtual_disk_size - Optional - Virtual Disk size in GB. Default 1GB.
  * virtual_disk_type - Optional - Virtual Disk type.  (thin, zeroedthick or eagerzeroedthick) Default 'thin'.


* resource "esxi_guest"
  * guest_name - Required - The Guest name.
  * ip_address - Computed - The IP address reported by VMware tools.
  * boot_disk_type - Optional - Guest boot disk type. Default 'thin'.  Available thin, zeroedthick, eagerzeroedthick.
  * boot_disk_size - Optional - Specify boot disk size or grow cloned vm to this size.
  * guestos - Optional - Default will be taken from cloned source.
  * boot_firmware - Optional - If "efi", enable efi boot. - Default "bios" (BIOS boot)
  * clone_from_vm - Source vm to clone. Mutually exclusive with ovf_source option.
  * ovf_source - ovf files or URLs to use as a source. Mutually exclusive with clone_from_vm option.
  * disk_store - Required - esxi Disk Store where guest vm will be created.
  * resource_pool_name - Optional - Any existing or terraform managed resource pool name. - Default "/".
  * memsize - Optional - Memory size in MB.  (ie, 1024 == 1GB). See esxi documentation for limits. - Default 512 or default taken from cloned source.
  * numvcpus - Optional - Number of virtual cpus.  See esxi documentation for limits. - Default 1 or default taken from cloned source.
  * virthwver - Optional - esxi guest virtual HW version.  See esxi documentation for compatible values. - Default 8 or taken from cloned source.
  * network_interfaces - Array of up to 10 network interfaces.
    * virtual_network - Required for each Guest NIC - This is the esxi virtual network name configured on esxi host.
    * mac_address - Optional -  If not set, mac_address will be generated by esxi.  Be sure to follow VMware mac address rules, otherwise your VM will not start.
    * nic_type - Optional - See esxi documentation for compatibility list. - Default "e1000" or taken from cloned source.
  * virtual_disks - Optional - Array of additional storage to be added to the guest.
    * virtual_disk_id - Required - virtual_disk.id from esxi_virtual_disk resource.
    * slot - Required - SCSI_Ctrl:SCSI_id.  Range  '0:1' to '3:15'.  SCSI_id 7 is not allowed.
  * power - Optional - on, off.
  * guest_startup_timeout - Optional - The amount of guest uptime, in seconds, to wait for an available IP address on this virtual machine. Default 120s.
  * guest_shutdown_timeout - Optional - The amount of time, in seconds, to wait for a graceful shutdown before doing a forced power off. Default 20s.
  * notes - Optional - The Guest notes (annotation).
  * guestinfo - Optional - The Guestinfo root
    * metadata - Optional - A JSON string containing the cloud-init metadata.
    * metadata.encoding - Optional - The encoding type for guestinfo.metadata. (base64 or gzip+base64)
    * userdata - Optional - A YAML document containing the cloud-init user data.
    * userdata.encoding - Optional - The encoding type for guestinfo.userdata. (base64 or gzip+base64)
    * vendordata - Optional - A YAML document containing the cloud-init vendor data.
    * vendordata.encoding - Optional - The encoding type for guestinfo.vendordata (base64 or gzip+base64)
  * ovf_properties - Optional - List of ovf properties to override in ovf/ova sources.
    * key - Required - Key of the property
    * value - Required - Value of the property
  * ovf_properties_timer - Optional - Length of time to wait for ovf_properties to process.  Default 90s.


* resource "esxi_vswitch"
  * name - Required - The vswitch name.
  * ports - Optional - The number of ports available on the vswitch.  Default 128.
  * mtu - Optional - The mtu. Default 1500
  * promiscuous_mode - Optional - Enable Promiscuous Mode (true/false) - Default false
  * mac_changes - Optional - Enable MAC Changes (true/false) - Default false.
  * forged_transmits - Optional - Enable Forged Transmits (true/false) - Default false.
  * uplink - Optional - Array of up to 32 uplinks.
    * name - Required - The uplink name. (for example vnic2)


* resource "esxi_portgroup"
  * name - Required - The portgroup name.    
  * vswitch - Required - The vswitch to connect to.
  * vlan - Optional - The vlan id of the portgroup - Default 0.


Using ovf_source & clone_from_vm
--------------------------------
* clone_from_vm clones from sources on the esxi host.
  * The source VM must be powered off.
  * If the source VM is stored in a resource group, you must specify the path, for example.
    * clone_from_vm = "my_resource_group/my_source_vm"
* ovf_source clones from sources on your local hard disk or a URL.
  * A local ova, ovf, vmx file.  See known issues (below) with vmx sources.
  * URL specifying a remote ova.
    * For example, Ubuntu cloud-images: https://cloud-images.ubuntu.com/trusty/current/trusty-server-cloudimg-amd64.ova
* If neither is specified, then a bare-metal VM will be created.  There will be no OS on this vm.  If the VM is powered on, it will default to a network PXE boot.  
* ovf_source & clone_from_vm are mutually exclusive.


Known issues with vmware_esxi
-----------------------------
* Using a local source vmx files should not have any networks configured.  There is very limited network interface mapping abilities in ovf_tools for vmx files.  It's best to simply clean out all network information from your vmx file.  The plugin will add network configuration to the destination vm guest as required.
* terraform import cannot import the guest disk type (thick, thin, etc) if the VM is powered on and cannot import the guest ip_address if it's powered off.
* Only numvcpus are supported.   numcores is not supported.
* Doesn't support CDrom or floppy.
* Doesn't support Shared bus Interfaces, or Shared disks
* Using an incorrect password could lockout your account using default esxi pam settings.
* Don't set guest_startup_timeout or guest_shutdown_timeout to 0 (zero).  It's valid, however it will be changed to default values by terraform.

Donations
---------
I work very hard to produce a stable, well documented product.  I appreciate any payments or donations for my efforts.
* Bitcoin: 1Kt89337143SzLjSddkRDVEMBRUWoKQhqy
* paypal:  josenk at jintegrate.co


Version History
---------------
* 1.10.3 Fix, Reload after disk expansion.  Disk exansion of IDE boot disks.
* 1.10.2 Fix to allow virtual_disk_dir to contain a '/'. (more then a single subdir depth)
* 1.10.0 Add support for boot_firmware. Fix portgroup to default to inherit from vSwitch.
* 1.9.1 Fix, Set default ovf_properties_timer.  Fix and add more details to example 06.
* 1.9.0 Manage portgroup security policies, fix typos.
* 1.8.3 Add support for ldap integrated esxi systems.
* 1.8.2 Fix, Disk Stores containing spaces for bare-metal builds.
* 1.8.1 Fix, multimachine create on Windows.
* 1.8.0 Add support for vswitch and portgroup resources.
* 1.7.2 Correctly set numvcpu when using clone_from_vm that doesn't have a numvcpu key set.
* 1.7.1 Bump release to include support for Terraform 0.13
* 1.7.0 Add support for esxi_hostssl (Setting the ssl port for ovftool).
* 1.6.4 Fix default IP dection. Fix Disk Stores containing spaces.
* 1.6.3 Mask username/password in debug logs.  Set default, disk.EnableUUID = true.
* 1.6.2 Fix Defaults for guest_startup_timeout and guest_shutdown_timeout.  Fix IP address detection type2 to always run regardless of guest_startup_timeout value.
* 1.6.1 Fix some minor refresh bugs, allow http(s) ovf sources.
* 1.6.0 Add support for ovf_properties for OVF/OVA sources.
* 1.5.4 Fix bare-metal build when using additional virtual disks.
* 1.5.3 Fix introduced bug when creating a bare-metal guest.
* 1.5.2 Handle large userdata using scp.  Connectivity test will retry only 3 times to help prevent account lockout.
* 1.5.1 Windows Fix for special characters in esxi password.
* 1.5.0 Support for Terraform 0.12, migrated examples to 0.12 format. Support to modify virtual_network & nic_type.  Windows fixes.
* 1.4.3 Fix virtdisk count. Fixes to support Terraform 0.12
* 1.4.2 Support 10 nics, more README changes
* 1.4.1 Fix README build instructions, static binaries, update guest types
* 1.4.0 Add GuestInfo (Cloud-init, Ignition!).   Fix, allow esxi passwords with special characters.
* 1.3.0 Add support to Update storage attachments.
* 1.2.2 fix guest_update power, boot_disk_type defaults, README, windows support
* 1.2.1 Fix ssh connection retries.
* 1.2.0 Add support for notes (annotation)
* 1.1.1 Fix, unable to provision ova sources.  go fmt.
* 1.1.0 Add Import support.
* 1.0.2 Switch authentication method to Keyboard Interactive.  Read disk_type (thin, thick, etc)
* 1.0.1 Validate DiskStores and refresh
* 1.0.0 First Major release
* 0.1.2 Add ability to manage existing Guest VMs.  A lot of code cleanup, various fixes, more validation.
* 0.1.0 Add virtual_disk resource.
* 0.0.8 Add virthwver.
* 0.0.7 build vmx from scratch if no source is specified
* 0.0.6 Add power resource.
* 0.0.5 Add network_interfaces resource.
* 0.0.4 Add more stuff.
* 0.0.3 Add memory and numvcpus resource.  Add support to update some guests params.
* 0.0.2 Add Resource Pool resource.
* 0.0.1 Init release
