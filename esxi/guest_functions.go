package esxi

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func guestGetVMID(c *Config, guest_name string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestGetVMID]\n")

	var remote_cmd, vmid string
	var err error

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null |sort -n | "+
		"grep -m 1 \"[0-9] * %s .*%s\" |awk '{print $1}' ", guest_name, guest_name)

	vmid, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "get vmid")
	log.Printf("[guestGetVMID] result: %s\n", vmid)
	if err != nil {
		log.Printf("[guestGetVMID] Failed get vmid: %s\n", err)
		return "", fmt.Errorf("Failed get vmid: %s\n", err)
	}

	return vmid, nil
}

func guestValidateVMID(c *Config, vmid string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestValidateVMID]\n")

	var remote_cmd string
	var err error

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | awk '{print $1}' | "+
		"grep '^%s$'", vmid)

	vmid, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "validate vmid exists")
	log.Printf("[guestValidateVMID] result: %s\n", vmid)
	if err != nil {
		log.Printf("[guestValidateVMID] Failed get vmid: %s\n", err)
		return "", fmt.Errorf("Failed get vmid: %s\n", err)
	}

	return vmid, nil
}

func getBootDiskPath(c *Config, vmid string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[getBootDiskPath]\n")

	var remote_cmd, stdout string
	var err error

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/device.getdevices %s | grep -A10 'key = 2000'|grep -m 1 fileName", vmid)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "get boot disk")
	if err != nil {
		log.Printf("[getBootDiskPath] Failed get boot disk path: %s\n", stdout)
		return "Failed get boot disk path:", err
	}
	r := strings.NewReplacer("fileName = \"[", "/vmfs/volumes/",
		"] ", "/", "\",", "")
	return r.Replace(stdout), err
}

func getDst_vmx_file(c *Config, vmid string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[getDst_vmx_file]\n")

	var dst_vmx_ds, dst_vmx, dst_vmx_file string

	//      -Get location of vmx file on esxi host
	remote_cmd := fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"", vmid)
	stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "get dst_vmx_ds")
	dst_vmx_ds = stdout
	dst_vmx_ds = strings.Trim(dst_vmx_ds, "[")
	dst_vmx_ds = strings.Trim(dst_vmx_ds, "]")

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'", vmid)
	stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "get dst_vmx")
	dst_vmx = stdout

	dst_vmx_file = "/vmfs/volumes/" + dst_vmx_ds + "/" + dst_vmx
	return dst_vmx_file, err
}

func readVmx_contents(c *Config, vmid string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[getVmx_contents]\n")

	var remote_cmd, vmx_contents string

	dst_vmx_file, err := getDst_vmx_file(c, vmid)
	remote_cmd = fmt.Sprintf("cat \"%s\"", dst_vmx_file)
	vmx_contents, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "read guest_name.vmx file")

	return vmx_contents, err
}

func updateVmx_contents(c *Config, vmid string, iscreate bool, memsize int, numvcpus int,
	virthwver int, guestos string, virtual_networks [10][4]string, virtual_disks [60][2]string, notes string,
	guestinfo map[string]interface{}) error {

	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[updateVmx_contents]\n")

	var regexReplacement, remote_cmd string

	vmx_contents, err := readVmx_contents(c, vmid)
	if err != nil {
		log.Printf("[updateVmx_contents] Failed get vmx contents: %s\n", err)
		return fmt.Errorf("Failed to get vmx contents: %s\n", err)
	}
	if strings.Contains(vmx_contents, "Unable to find a VM corresponding") {
		return nil
	}

	// modify memsize
	if memsize != 0 {
		re := regexp.MustCompile("memSize = \".*\"")
		regexReplacement = fmt.Sprintf("memSize = \"%d\"", memsize)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify numvcpus
	if numvcpus != 0 {
		if strings.Contains(vmx_contents, "numvcpus = ") {
			re := regexp.MustCompile("numvcpus = \".*\"")
			regexReplacement = fmt.Sprintf("numvcpus = \"%d\"", numvcpus)
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
		} else {
			log.Printf("[updateVmx_contents] Add numvcpu: %d\n", numvcpus)
			vmx_contents += fmt.Sprintf("\nnumvcpus = \"%d\"", numvcpus)
		}
	}

	// modify virthwver
	if virthwver != 0 {
		re := regexp.MustCompile("virtualHW.version = \".*\"")
		regexReplacement = fmt.Sprintf("virtualHW.version = \"%d\"", virthwver)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify guestos
	if guestos != "" {
		re := regexp.MustCompile("guestOS = \".*\"")
		regexReplacement = fmt.Sprintf("guestOS = \"%s\"", guestos)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify annotation
	if notes != "" {
		notes = strings.Replace(notes, "\"", "|22", -1)
		if strings.Contains(vmx_contents, "annotation") {
			re := regexp.MustCompile("annotation = \".*\"")
			regexReplacement = fmt.Sprintf("annotation = \"%s\"", notes)
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
		} else {
			regexReplacement = fmt.Sprintf("\nannotation = \"%s\"", notes)
			vmx_contents += regexReplacement
		}
	}

	if len(guestinfo) > 0 {
		parsed_vmx := ParseVMX(vmx_contents)
		for k, v := range guestinfo {
			log.Println("SAVING", k, v)
			parsed_vmx["guestinfo."+k] = v.(string)
		}
		vmx_contents = EncodeVMX(parsed_vmx)
	}

	//
	//  add/modify virtual disks
	//
	var tmpvar string
	var vmx_contents_new string
	var i, j int

	//
	//  Remove all disks
	//
	regexReplacement = fmt.Sprintf("")
	for i = 0; i < 4; i++ {
		for j = 0; j < 16; j++ {

			if (i != 0 || j != 0) && j != 7 {
				re := regexp.MustCompile(fmt.Sprintf("scsi%d:%d.*\n", i, j))
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
			}
		}
	}

	//
	//  Add disks that are managed by terraform
	//
	for i = 0; i < 59; i++ {
		if virtual_disks[i][0] != "" {

			log.Printf("[updateVmx_contents] Adding: %s\n", virtual_disks[i][1])
			tmpvar = fmt.Sprintf("scsi%s.deviceType = \"scsi-hardDisk\"\n", virtual_disks[i][1])
			if !strings.Contains(vmx_contents, tmpvar) {
				vmx_contents += "\n" + tmpvar
			}

			tmpvar = fmt.Sprintf("scsi%s.fileName", virtual_disks[i][1])
			if strings.Contains(vmx_contents, tmpvar) {
				re := regexp.MustCompile(tmpvar + " = \".*\"")
				regexReplacement = fmt.Sprintf(tmpvar+" = \"%s\"", virtual_disks[i][0])
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
			} else {
				regexReplacement = fmt.Sprintf("\n"+tmpvar+" = \"%s\"", virtual_disks[i][0])
				vmx_contents += "\n" + regexReplacement
			}

			tmpvar = fmt.Sprintf("scsi%s.present = \"true\"\n", virtual_disks[i][1])
			if !strings.Contains(vmx_contents, tmpvar) {
				vmx_contents += "\n" + tmpvar
			}

		}
	}

	//
	//  Create/update networks network_interfaces
	//

	//  Define default nic type.
	var defaultNetworkType, networkType string
	if virtual_networks[0][2] != "" {
		defaultNetworkType = virtual_networks[0][2]
	} else {
		defaultNetworkType = "e1000"
	}

	//  If this is first time provisioning, delete all the old ethernet configuration.
	if iscreate == true {
		log.Printf("[updateVmx_contents] Delete old ethernet configuration: %d\n", i)
		regexReplacement = fmt.Sprintf("")
		for i = 0; i < 9; i++ {
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
		}
	}

	//  Add/Modify virtual networks.
	networkType = ""

	for i := 0; i <= 9; i++ {
		log.Printf("[updateVmx_contents] ethernet%d\n", i)

		if virtual_networks[i][0] == "" && strings.Contains(vmx_contents, "ethernet"+strconv.Itoa(i)) == true {
			//  This is Modify (Delete existing network configuration)
			log.Printf("[updateVmx_contents] Modify ethernet%d - Delete existing.\n", i)
			regexReplacement = fmt.Sprintf("")
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
		}

		if virtual_networks[i][0] != "" && strings.Contains(vmx_contents, "ethernet"+strconv.Itoa(i)) == true {
			//  This is Modify
			log.Printf("[updateVmx_contents] Modify ethernet%d - Modify existing.\n", i)

			//  Modify Network Name
			re := regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".networkName = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".networkName = \"%s\"", virtual_networks[i][0])
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)

			//  Modify virtual Device
			re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".virtualDev = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".virtualDev = \"%s\"", virtual_networks[i][2])
			vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)

			//  Modify MAC (dynamic to static only. static to dynamic is not implemented)
			if virtual_networks[i][1] != "" {
				log.Printf("[updateVmx_contents] ethernet%d Modify MAC: %s\n", i, virtual_networks[i][0])

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".[a-zA-Z]*ddress = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".address = \"%s\"", virtual_networks[i][1])
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".addressType = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet" + strconv.Itoa(i) + ".addressType = \"static\"")
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".generatedAddressOffset = \".*\"")
				regexReplacement = fmt.Sprintf("")
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
			}
		}

		if virtual_networks[i][0] != "" && strings.Contains(vmx_contents, "ethernet"+strconv.Itoa(i)) == false {
			//  This is create

			//  Set virtual_network name
			log.Printf("[updateVmx_contents] ethernet%d Create New: %s\n", i, virtual_networks[i][0])
			tmpvar = fmt.Sprintf("\nethernet%d.networkName = \"%s\"\n", i, virtual_networks[i][0])
			vmx_contents_new = tmpvar

			//  Set mac address
			if virtual_networks[i][1] != "" {
				tmpvar = fmt.Sprintf("ethernet%d.addressType = \"static\"\n", i)
				vmx_contents_new = vmx_contents_new + tmpvar

				tmpvar = fmt.Sprintf("ethernet%d.address = \"%s\"\n", i, virtual_networks[i][1])
				vmx_contents_new = vmx_contents_new + tmpvar
			}

			//  Set network type
			if virtual_networks[i][2] == "" {
				networkType = defaultNetworkType
			} else {
				networkType = virtual_networks[i][2]
			}

			tmpvar = fmt.Sprintf("ethernet%d.virtualDev = \"%s\"\n", i, networkType)
			vmx_contents_new = vmx_contents_new + tmpvar

			tmpvar = fmt.Sprintf("ethernet%d.present = \"TRUE\"\n", i)

			vmx_contents = vmx_contents + vmx_contents_new + tmpvar
		}
	}

	//  Add disk UUID
	if !strings.Contains(vmx_contents, "disk.EnableUUID") {
		vmx_contents = vmx_contents + "\ndisk.EnableUUID = \"TRUE\""
	}

	//
	//  Write vmx file to esxi host
	//
	log.Printf("[updateVmx_contents] New guest_name.vmx: %s\n", vmx_contents)

	dst_vmx_file, err := getDst_vmx_file(c, vmid)

	vmx_contents, err = writeContentToRemoteFile(esxiConnInfo, strings.Replace(vmx_contents, "\\\"", "\"", -1), dst_vmx_file, "write guest_name.vmx file")

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/reload %s", vmid)
	_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/reload")
	return err
}

func cleanStorageFromVmx(c *Config, vmid string) error {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[cleanStorageFromVmx]\n")

	var remote_cmd string

	vmx_contents, err := readVmx_contents(c, vmid)
	if err != nil {
		log.Printf("[updateVmx_contents] Failed get vmx contents: %s\n", err)
		return fmt.Errorf("Failed to get vmx contents: %s\n", err)
	}

	for x := 0; x < 4; x++ {
		for y := 0; y < 16; y++ {
			if !(x == 0 && y == 0) {
				regexReplacement := fmt.Sprintf("scsi%d:%d.*", x, y)
				re := regexp.MustCompile(regexReplacement)
				vmx_contents = re.ReplaceAllString(vmx_contents, "")
			}
		}
	}

	//
	//  Write vmx file to esxi host
	//

	dst_vmx_file, err := getDst_vmx_file(c, vmid)
	vmx_contents, err = writeContentToRemoteFile(esxiConnInfo, strings.Replace(vmx_contents, "\\\"", "\"", -1), dst_vmx_file, "write guest_name.vmx file")
	remote_cmd = fmt.Sprintf("sed -i '/^$/d' %s", dst_vmx_file)

	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/reload %s", vmid)
	_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/reload")
	return err
}

func guestPowerOn(c *Config, vmid string) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestPowerOn]\n")

	if guestPowerGetState(c, vmid) == "on" {
		return "", nil
	}

	remote_cmd := fmt.Sprintf("vim-cmd vmsvc/power.on %s", vmid)
	stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/power.on")
	time.Sleep(3 * time.Second)

	if guestPowerGetState(c, vmid) == "on" {
		return stdout, nil
	}

	return stdout, err
}

func guestPowerOff(c *Config, vmid string, guest_shutdown_timeout int) (string, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestPowerOff]\n")

	var remote_cmd, stdout string

	savedpowerstate := guestPowerGetState(c, vmid)
	if savedpowerstate == "off" {
		return "", nil

	} else if savedpowerstate == "on" {

		if guest_shutdown_timeout != 0 {
			remote_cmd = fmt.Sprintf("vim-cmd vmsvc/power.shutdown %s", vmid)
			stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/power.shutdown")
			time.Sleep(3 * time.Second)

			for i := 0; i < (guest_shutdown_timeout / 3); i++ {
				if guestPowerGetState(c, vmid) == "off" {
					return stdout, nil
				}
				time.Sleep(3 * time.Second)
			}
		}

		remote_cmd = fmt.Sprintf("vim-cmd vmsvc/power.off %s", vmid)
		stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/power.off")
		time.Sleep(1 * time.Second)

		return stdout, nil

	} else {
		remote_cmd = fmt.Sprintf("vim-cmd vmsvc/power.off %s", vmid)
		stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/power.off")
		return stdout, nil
	}
}

func guestPowerGetState(c *Config, vmid string) string {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestPowerGetState]\n")

	remote_cmd := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s", vmid)
	stdout, _ := runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmsvc/power.getstate")
	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		return "Unknown"
	}

	if strings.Contains(stdout, "Powered off") == true {
		return "off"
	} else if strings.Contains(stdout, "Powered on") == true {
		return "on"
	} else if strings.Contains(stdout, "Suspended") == true {
		return "suspended"
	} else {
		return "Unknown"
	}
}

func guestGetIpAddress(c *Config, vmid string, guest_startup_timeout int) string {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestGetIpAddress]\n")

	var remote_cmd, stdout, ip_address, ip_address2 string
	var uptime int

	//  Check if powered off
	if guestPowerGetState(c, vmid) != "on" {
		return ""
	}

	//
	//  Check uptime of guest.
	//
	uptime = 0
	for uptime < guest_startup_timeout {
		//  Primary method to get IP
		remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null |sed '1!G;h;$!d' |awk '/deviceConfigId = 4000/,/ipAddress/' |grep -m 1 -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", vmid)
		stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "get ip_address method 1")
		ip_address = stdout
		if ip_address != "" {
			return ip_address
		}

		time.Sleep(3 * time.Second)

		//  Get uptime if above failed.
		remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null | grep 'uptimeSeconds ='|sed 's/^.*= //g'|sed s/,//g", vmid)
		stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "get uptime")
		if err != nil {
			return ""
		}
		uptime, _ = strconv.Atoi(stdout)
	}

	//
	// Alternate method to get IP
	//
	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null | grep -m 1 '^   ipAddress = ' | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", vmid)
	stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "get ip_address method 2")
	ip_address2 = stdout
	if ip_address2 != "" {
		return ip_address2
	}

	return ""
}
