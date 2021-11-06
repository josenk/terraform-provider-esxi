# Terraform esxi Provider (06 OVF Properties)
---

## Notes on ovf_properties.
* There are a few caveats using this method of configuring your guest.
  * This method of injecting properties into the VM works with ovf/ova sources only.  The guestinfo method ("05 CloudInit and Templates" example) can be used with any source type (vmx, ovf, ova, clone_from_vm), but requires "Cloud-Init for VMware" to be pre-installed on the source.

  * 'terraform apply' takes longer to deploy when using ovf_properties because the guest will be booted twice.  The default length of time to allow ovf_properties to run is 90 seconds.  If your configuration requires more time, set ovf_properties_timer to a higher value.
  * You cannot configure additional virtual_disks in userdata using the ovf_properties method.  Userdata in ovf_properties runs before the additional virtual_disks are added.  If this is a requirement, use the guestinfo (Cloud-Init for VMware) method.
  
  * This method only works on ovf/ova sources that have properties available to configure.  Not all ovf/ova files will have properties or their Key naming convention may not be consistent between sources.

## How to use ovf_properties.
* Some ovf/ova files have available properties.   These key/value properties can be retrieved from the ovf/ova file.  The default values of the guest can be over written using the ovf_properties configuration option in your terraform code.

* To list the OVF properties in a specific ovf/ova file, use the following command.

```
ovftool --hideEula image.ova
```

Example output

```
Properties:
  Key:         instance-id
  Label:       A Unique Instance ID for this instance
  Type:        string
  Description: Specifies the instance id.  This is required and used to
               determine if the machine should take "first boot" actions
  Value:       id-ovf

  Key:         hostname
  Type:        string
  Description: Specifies the hostname for the appliance
  Value:       ubuntuguest

  Key:         user-data
  Label:       Encoded user-data
  Type:        string
  Description: In order to fit into a xml attribute, this value is base64
               encoded . It will be decoded, and then processed normally as
               user-data.

  Key:         password
  Label:       Default User's password
  Type:        string
  Description: If set, the default user's password will be set to this value to
               allow password based login.  The password will be good for only
               a single login.  If set to the string 'RANDOM' then a random
               password will be generated, and written to the console.
  Value:       q

```

* The `user-data` key allows you to inject user-data via ovf_properties.   This works with images that have cloud-init included.
  * For example: Ubuntu cloud-images. https://cloud-images.ubuntu.com/

* The following example will set the "Default User's Password" and hostname.   It will also pass the rendered userdata.tpl file to the vm.
```
  ovf_properties {
    key = "password"
    value = "Passw0rd1"
  }

  ovf_properties {
    key = "hostname"
    value = "vmtest06"
  }

  ovf_properties {
    key = "user-data"
    value = base64encode(data.template_file.userdata_default.rendered)
  }
  ```
