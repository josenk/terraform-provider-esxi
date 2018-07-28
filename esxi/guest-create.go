package esxi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"errors"
)


func guestCREATE(c *Config, guest_name string, disk_store string,
	 src_path string, resource_pool_name string, memsize int, numvcpus int, virthwver int,
	 boot_disk_type string, boot_disk_size string, virtual_networks [4][3]string , virtual_disks [60][2]string) (string, error) {

  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestCREATE]\n")

  var boot_disk_vmdkPATH, remote_cmd, vmid, stdout, vmx_contents string
	var out bytes.Buffer
	var err error
	err = nil

  if src_path == "none" {

		// check if path already exists.
		fullPATH := fmt.Sprintf("/vmfs/volumes/%s/%s", disk_store, guest_name)
		boot_disk_vmdkPATH = fmt.Sprintf("/vmfs/volumes/%s/%s/%s.vmdk", disk_store, guest_name, guest_name)
    remote_cmd = fmt.Sprintf("ls -d %s", fullPATH)
		stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "check if guest path already exists.")
		if stdout == fullPATH {
			fmt.Println("[provider-esxi] guest path already exists: " + err.Error())
	  	return "Error", errors.New("Error: guest path already exists: " + err.Error())
		} else {
  		remote_cmd = fmt.Sprintf("mkdir %s", fullPATH)
	  	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create guest path")
			if err != nil {
				log.Println("[provider-esxi] Error creating guest path: " + err.Error())
		  	return err.Error(), err
			}
		}


		guestOS := "centos-64"
		hasISO := false
		isofilename := ""

		//if numvcpus == 0 {
		//	numvcpus = 1
		//}
    if memsize == 0 {
			memsize = 512
		}
		if virthwver == 0 {
			virthwver = 8
		}


		// Build VM by default/black config
    vmx_contents =
		  fmt.Sprintf("config.version = \\\"8\\\"\n") +
			fmt.Sprintf("virtualHW.version = \\\"%d\\\"\n", virthwver) +
			fmt.Sprintf("displayName = \\\"%s\\\"\n", guest_name) +
			fmt.Sprintf("numvcpus = \\\"%d\\\"\n", numvcpus) +
			fmt.Sprintf("memSize = \\\"%d\\\"\n", memsize) +
			fmt.Sprintf("guestOS = \\\"%s\\\"\n", guestOS) +
			fmt.Sprintf("floppy0.present = \\\"FALSE\\\"\n") +
			fmt.Sprintf("scsi0.present = \\\"TRUE\\\"\n") +
			fmt.Sprintf("scsi0.sharedBus = \\\"none\\\"\n") +
			fmt.Sprintf("scsi0.virtualDev = \\\"lsilogic\\\"\n") +
			fmt.Sprintf("pciBridge0.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("pciBridge4.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("pciBridge4.virtualDev = \\\"pcieRootPort\\\"\n") +
      fmt.Sprintf("pciBridge4.functions = \\\"8\\\"\n") +
      fmt.Sprintf("pciBridge5.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("pciBridge5.virtualDev = \\\"pcieRootPort\\\"\n") +
      fmt.Sprintf("pciBridge5.functions = \\\"8\\\"\n") +
      fmt.Sprintf("pciBridge6.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("pciBridge6.virtualDev = \\\"pcieRootPort\\\"\n") +
      fmt.Sprintf("pciBridge6.functions = \\\"8\\\"\n") +
      fmt.Sprintf("pciBridge7.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("pciBridge7.virtualDev = \\\"pcieRootPort\\\"\n") +
      fmt.Sprintf("pciBridge7.functions = \\\"8\\\"\n") +
			fmt.Sprintf("scsi0:0.present = \\\"TRUE\\\"\n") +
      fmt.Sprintf("scsi0:0.fileName = \\\"%s.vmdk\\\"\n", guest_name) +
      fmt.Sprintf("scsi0:0.deviceType = \\\"scsi-hardDisk\\\"\n")
		if hasISO == true {
			vmx_contents = vmx_contents +
				fmt.Sprintf("ide1:0.present = \\\"TRUE\\\"\n") +
        fmt.Sprintf("ide1:0.fileName = \\\"emptyBackingString\\\"\n") +
        fmt.Sprintf("ide1:0.deviceType = \\\"atapi-cdrom\\\"\n") +
        fmt.Sprintf("ide1:0.startConnected = \\\"FALSE\\\"\n") +
        fmt.Sprintf("ide1:0.clientDevice = \\\"TRUE\\\"\n")
		} else {
		  vmx_contents = vmx_contents +
		    fmt.Sprintf("ide1:0.present = \\\"TRUE\\\"\n") +
	      fmt.Sprintf("ide1:0.fileName = \\\"%s\\\"\n", isofilename) +
	      fmt.Sprintf("ide1:0.deviceType = \\\"cdrom-image\\\"\n")
		}

		//
		//  Write vmx file to esxi host
		//
		log.Printf("[provider-esxi] New guest_name.vmx: %s\n", vmx_contents)

		dst_vmx_file := fmt.Sprintf("%s/%s.vmx", fullPATH, guest_name)

		remote_cmd = fmt.Sprintf("echo \"%s\" >%s", vmx_contents, dst_vmx_file)
		vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "write guest_name.vmx file")

		//  Create boot disk (vmdk)
		remote_cmd  = fmt.Sprintf("vmkfstools -c %sG -d %s %s/%s.vmdk", boot_disk_size, boot_disk_type, fullPATH, guest_name)
		_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmkfstools (make boot disk)")
		if err != nil {
			remote_cmd = fmt.Sprintf("rm -fr %s", fullPATH)
	  	stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "cleanup guest path because of failed events")
			log.Println("[provider-esxi] unable to vmkfstools (make boot disk): " + err.Error())
			return err.Error(), err
		}

    poolID, err := getPoolID(c, resource_pool_name)
		log.Println("[provider-esxi] DEBUG: " + poolID)
		if err != nil {
			log.Println("[provider-esxi] unable to Resource Pool ID: " + err.Error())
			return err.Error(), err
		}
		remote_cmd  = fmt.Sprintf("vim-cmd solo/registervm %s %s %s",dst_vmx_file, guest_name, poolID)
		_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "solo/registervm")
		if err != nil {
			log.Println("[provider-esxi] unable to register guest: " + err.Error())
			remote_cmd = fmt.Sprintf("rm -fr %s", fullPATH)
	  	stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "cleanup guest path because of failed events")
			return err.Error(), err
		}


	} else {
	  //  Build VM by ovftool
	  dst_path := fmt.Sprintf("vi://%s:%s@%s/%s", c.Esxi_username, c.Esxi_password, c.Esxi_hostname, resource_pool_name)

	  ovf_cmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false " +
	  	"-dm=%s --name='%s' --overwrite -ds='%s' '%s' '%s'",boot_disk_type, guest_name, disk_store, src_path, dst_path)
	  cmd := exec.Command("/bin/bash", "-c", ovf_cmd)

    log.Println("[provider-esxi] ovf_cmd: " + ovf_cmd )

	  cmd.Stdout = &out
	  err = cmd.Run()
	  log.Printf("[provider-esxi] ovftool output: %q\n", out.String())
	  if err != nil {
	  	log.Println("[provider-esxi] There was an ovftool Error: " + err.Error())
	  	return err.Error(), err
	  }

  }

  remote_cmd = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | sort -n | " +
		"grep \"[0-9] * %s .*%s\" | awk '{print $1}' | " +
		"tail -1", guest_name, guest_name)

  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get vmid")
	vmid = stdout
	log.Printf("[provider-esxi] get_vmid_cmd: %s\n", vmid)
	if err != nil {
		log.Printf("[provider-esxi] Failed get vmid_cmd: %s\n", stdout)
		return "Failed get vmid", err
	}

	//
	//  Grow boot disk to boot_disk_size
	//
	remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/device.getdevices %s | grep -A10 'key = 2000'|grep -m 1 fileName", vmid)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get boot disk")
	if err != nil {
		log.Printf("[provider-esxi] Failed get boot disk path: %s\n", stdout)
		return "Failed get boot disk path:", err
	}
	r := strings.NewReplacer("fileName = \"[", "/vmfs/volumes/",
													 "] ", "/", "\",", "")
	boot_disk_vmdkPATH = r.Replace(stdout)
	log.Printf("[provider-esxi] fullPATH: %s\n", boot_disk_vmdkPATH)

	if boot_disk_size != "" {
		err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
		if err != nil {
			return vmid, errors.New("Unable to grow boot disk.")
		}
	}

	//
	//  make updates to vmx file
	//
  err = updateVmx_contents(c, vmid, true, memsize, numvcpus, virthwver, virtual_networks, virtual_disks)
	if err != nil {
		return vmid, err
	}

  return vmid, err
}
