package esxi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

func guestCREATE(c *Config, guest_name string, disk_store string,
	src_path string, resource_pool_name string, strmemsize string, strnumvcpus string, strvirthwver string, guestos string,
	boot_disk_type string, boot_disk_size string, virtual_networks [4][3]string,
	virtual_disks [60][2]string, guest_shutdown_timeout int) (string, error) {
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Printf("[guestCREATE]\n")

	var memsize, numvcpus, virthwver int
	var boot_disk_vmdkPATH, remote_cmd, vmid, stdout, vmx_contents string
	var out bytes.Buffer
	var err error
	err = nil

	memsize, _ = strconv.Atoi(strmemsize)
	numvcpus, _ = strconv.Atoi(strnumvcpus)
	virthwver, _ = strconv.Atoi(strvirthwver)

	//
	//  Check if Disk Store already exists
	//
	err = diskStoreValidate(c, disk_store)
	if err != nil {
		return "", err
	}

	//
	//  Check if guest already exists
	//
	// get VMID (by name)
	vmid, err = guestGetVMID(c, guest_name)

	if vmid != "" {
		// We don't need to create the VM.   It already exists.
		fmt.Printf("[guestCREATE] guest %s already exists vmid: \n", guest_name, stdout)

		//
		//   Power off guest if it's powered on.
		//
		currentpowerstate := guestPowerGetState(c, vmid)
		if currentpowerstate == "on" || currentpowerstate == "suspended" {
			_, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
			if err != nil {
				return "", fmt.Errorf("Failed to power off existing guest. vmid:%s\n", vmid)
			}
		}

	} else if src_path == "none" {

		// check if path already exists.
		fullPATH := fmt.Sprintf("\"/vmfs/volumes/%s/%s\"", disk_store, guest_name)
		boot_disk_vmdkPATH = fmt.Sprintf("\"/vmfs/volumes/%s/%s/%s.vmdk\"", disk_store, guest_name, guest_name)
		remote_cmd = fmt.Sprintf("ls -d %s", fullPATH)
		stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "check if guest path already exists.")
		if stdout == fullPATH {
			fmt.Printf("Error: Guest path already exists. fullPATH:%s\n", fullPATH)
			return "", fmt.Errorf("Guest path already exists. fullPATH:%s\n", fullPATH)
		} else {
			remote_cmd = fmt.Sprintf("mkdir %s", fullPATH)
			stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create guest path")
			if err != nil {
				log.Printf("Failed to create guest path. fullPATH:%s\n", fullPATH)
				return "", fmt.Errorf("Failed to create guest path. fullPATH:%s\n", fullPATH)
			}
		}

		hasISO := false
		isofilename := ""

		if numvcpus == 0 {
			numvcpus = 1
		}
		if memsize == 0 {
			memsize = 512
		}
		if virthwver == 0 {
			virthwver = 8
		}
		if guestos == "" {
			guestos = "centos-64"
		}
		if boot_disk_size == "" {
			boot_disk_size = "16"
		}

		// Build VM by default/black config
		vmx_contents =
			fmt.Sprintf("config.version = \\\"8\\\"\n") +
				fmt.Sprintf("virtualHW.version = \\\"%d\\\"\n", virthwver) +
				fmt.Sprintf("displayName = \\\"%s\\\"\n", guest_name) +
				fmt.Sprintf("numvcpus = \\\"%d\\\"\n", numvcpus) +
				fmt.Sprintf("memSize = \\\"%d\\\"\n", memsize) +
				fmt.Sprintf("guestOS = \\\"%s\\\"\n", guestos) +
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
		log.Printf("[guestCREATE] New guest_name.vmx: %s\n", vmx_contents)

		dst_vmx_file := fmt.Sprintf("%s/%s.vmx", fullPATH, guest_name)

		remote_cmd = fmt.Sprintf("echo \"%s\" >%s", vmx_contents, dst_vmx_file)
		vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "write guest_name.vmx file")

		//  Create boot disk (vmdk)
		remote_cmd = fmt.Sprintf("vmkfstools -c %sG -d %s %s/%s.vmdk", boot_disk_size, boot_disk_type, fullPATH, guest_name)
		_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmkfstools (make boot disk)")
		if err != nil {
			remote_cmd = fmt.Sprintf("rm -fr %s", fullPATH)
			stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "cleanup guest path because of failed events")
			log.Printf("Failed to vmkfstools (make boot disk):%s\n", err.Error())
			return "", fmt.Errorf("Failed to vmkfstools (make boot disk):%s\n", err.Error())
		}

		poolID, err := getPoolID(c, resource_pool_name)
		log.Println("[guestCREATE] DEBUG: " + poolID)
		if err != nil {
			log.Printf("Failed to use Resource Pool ID:%s\n", poolID)
			return "", fmt.Errorf("Failed to use Resource Pool ID:%s\n", poolID)
		}
		remote_cmd = fmt.Sprintf("vim-cmd solo/registervm %s %s %s", dst_vmx_file, guest_name, poolID)
		_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "solo/registervm")
		if err != nil {
			log.Printf("Failed to register guest:%s\n", err.Error())
			remote_cmd = fmt.Sprintf("rm -fr %s", fullPATH)
			stdout, _ = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "cleanup guest path because of failed events")
			return "", fmt.Errorf("Failed to register guest:%s\n", err.Error())
		}

	} else {
		//  Build VM by ovftool
		if boot_disk_type == "zeroedthick" {
			boot_disk_type = "thick"
		}
		dst_path := fmt.Sprintf("vi://%s:%s@%s/%s", c.esxiUserName, c.esxiPassword, c.esxiHostName, resource_pool_name)

		ovf_cmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false "+
			"-dm=%s --name='%s' --overwrite -ds='%s' '%s' '%s'", boot_disk_type, guest_name, disk_store, src_path, dst_path)
		cmd := exec.Command("/bin/bash", "-c", ovf_cmd)

		log.Println("[guestCREATE] ovf_cmd: " + ovf_cmd)

		cmd.Stdout = &out
		err = cmd.Run()
		log.Printf("[guestCREATE] ovftool output: %q\n", out.String())
		if err != nil {
			log.Printf("Failed, There was an ovftool Error:%s\n", err.Error())
			return "", fmt.Errorf("There was an ovftool Error:%s\n", err.Error())
		}

	}

	// get VMID (by name)
	vmid, err = guestGetVMID(c, guest_name)
	if err != nil {
		return "", err
	}

	//
	//  Grow boot disk to boot_disk_size
	//
	boot_disk_vmdkPATH, _ = getBootDiskPath(c, vmid)

	err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
	if err != nil {
		return vmid, fmt.Errorf("Failed to grow boot disk.\n")
	}

	//
	//  make updates to vmx file
	//
	err = updateVmx_contents(c, vmid, true, memsize, numvcpus, virthwver, guestos, virtual_networks, virtual_disks)
	if err != nil {
		return vmid, err
	}

	return vmid, nil
}
