package esxi

import (
	"fmt"
	"strings"
	"strconv"
	"log"
  "regexp"
	"bufio"
	"time"
)

func getDst_vmx_file(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getDst_vmx_file]\n")

  //      -Get location of vmx file on esxi host
  var dst_vmx_ds, dst_vmx, dst_vmx_file string

  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"",vmid)
	stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx_ds")
	dst_vmx_ds   = stdout
	dst_vmx_ds   = strings.Trim(dst_vmx_ds, "[")
	dst_vmx_ds   = strings.Trim(dst_vmx_ds, "]")

	remote_cmd   = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'",vmid)
	stdout, err  = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx")
	dst_vmx      = stdout

	dst_vmx_file = "/vmfs/volumes/" + dst_vmx_ds + "/" + dst_vmx
  return dst_vmx_file, err
}

func readVmx_contents(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getVmx_contents]\n")

  var remote_cmd, vmx_contents string

  dst_vmx_file,err := getDst_vmx_file(c, vmid)
  remote_cmd = fmt.Sprintf("cat %s", dst_vmx_file)
  vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "read guest_name.vmx file")

  return vmx_contents, err
}


func updateVmx_contents(c *Config, vmid string, iscreate bool, memsize int, numvcpus int,
	virthwver int, virtual_networks [4][3]string, virtual_disks [60][2]string) error {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / updateVmx_contents]\n")

  var regexReplacement, remote_cmd string

  vmx_contents, err := readVmx_contents(c, vmid)
	if err != nil {
		log.Printf("[provider-esxi] Failed get vmx contents: %s\n", err)
		return err
	}

	// modify memsize
  if memsize != 0 {
		re := regexp.MustCompile("memSize = \".*\"")
		regexReplacement = fmt.Sprintf("memSize = \"%d\"", memsize)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify numvcpus
	if numvcpus != 0 {
		re := regexp.MustCompile("numvcpus = \".*\"")
		regexReplacement = fmt.Sprintf("numvcpus = \"%d\"", numvcpus)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify virthwver
	if virthwver != 0 {
		re := regexp.MustCompile("virtualHW.version = \".*\"")
		regexReplacement = fmt.Sprintf("virtualHW.version = \"%d\"", virthwver)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

  //
	//  Create/update networks network_interfaces
	//
	var tmpvar string
	var vmx_contents_new string

	if iscreate == true {

		//  This is create network interfaces.  Clean out old network interfaces.
	  scanner := bufio.NewScanner(strings.NewReader(vmx_contents))
    for scanner.Scan() {

	  	if  scanner.Text() == "" || strings.Contains(scanner.Text(),"ethernet") == true {
	  		// Do nothing
				log.Printf("%s: skipped\n", scanner.Text())
			} else {
	  		vmx_contents_new = vmx_contents_new + scanner.Text() + "\n"
	  	}
	  }

    //  Add virtual networks.
		var defaultNetworkType, networkType string
		if virtual_networks[0][2] != "" {
		  defaultNetworkType = virtual_networks[0][2]
		} else {
			defaultNetworkType = "e1000"
		}
		networkType = ""

	  for i := 0; i < 4; i++ {
	  	log.Printf("[provider-esxi] i: %s\n", i)

	  	if virtual_networks[i][0] != "" {

				//  Set virtual_network name
	  		log.Printf("[provider-esxi] virtual_networks[i][0]: %s\n", virtual_networks[i][0])
	  		tmpvar = fmt.Sprintf("ethernet%d.networkName = \"%s\"\n", i, virtual_networks[i][0])
	  		vmx_contents_new = vmx_contents_new + tmpvar

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

	  		vmx_contents_new = vmx_contents_new + tmpvar
	  	}
	  }

		//  add virtual disks
		for i := 0; i < 59; i++ {
			if virtual_disks[i][0] != "" {
				tmpvar = fmt.Sprintf("scsi%s.deviceType = \"scsi-hardDisk\"\n", virtual_disks[i][1])
				vmx_contents_new = vmx_contents_new + tmpvar

				tmpvar = fmt.Sprintf("scsi%s.fileName = \"%s\"\n", virtual_disks[i][1], virtual_disks[i][0])
				vmx_contents_new = vmx_contents_new + tmpvar

				tmpvar = fmt.Sprintf("scsi%s.present = \"true\"\n", virtual_disks[i][1])
				vmx_contents_new = vmx_contents_new + tmpvar
			}
		}

		//  Save
		vmx_contents = vmx_contents_new

	} else {

		//  This is modify network interfaces
		for i := 0; i < 4; i++ {

			// Fix virtual_network
			if virtual_networks[i][0] != "" {
				re := regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".networkName = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet" + strconv.Itoa(i) + ".networkName = \"%s\"", virtual_networks[i][0])
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
			}

      //  Fix device type
			if virtual_networks[i][0] != "" && virtual_networks[i][2] != "" {
				re := regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".virtualDev = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet" + strconv.Itoa(i) + ".virtualDev = \"%s\"", virtual_networks[i][2])
				vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
			}
		}
	}


	//
	//  Write vmx file to esxi host
	//
	vmx_contents = strings.Replace(vmx_contents, "\"", "\\\"", -1)
	log.Printf("[provider-esxi] New guest_name.vmx: %s\n", vmx_contents)

  dst_vmx_file,err := getDst_vmx_file(c, vmid)
  remote_cmd = fmt.Sprintf("echo \"%s\" >%s", vmx_contents, dst_vmx_file)
	vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "write guest_name.vmx file")

	remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/reload %s",vmid)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/reload")
  return err
}


func guestPowerOn(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestPowerOn]\n")

	if guestPowerGetState(c, vmid) == "on" {
		return "",nil
	}

  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.on %s",vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.on")
	time.Sleep(3 * time.Second)

	if guestPowerGetState(c, vmid) == "on" {
		return stdout,nil
	}

  return stdout,err
}

func guestPowerOff(c *Config, vmid string, guest_shutdown_timeout int) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestPowerOff]\n")

  var remote_cmd, stdout string

  savedpowerstate := guestPowerGetState(c, vmid)
	if savedpowerstate == "off" {
		return "",nil

	} else if savedpowerstate == "on" {

		if guest_shutdown_timeout != 0 {
	    remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/power.shutdown %s",vmid)
	    stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.shutdown")
	    time.Sleep(3 * time.Second)

	    for i := 0; i < (guest_shutdown_timeout / 3); i++ {
	    	if guestPowerGetState(c, vmid) == "off" {
	    		return stdout,nil
	    	}
	    	time.Sleep(3 * time.Second)
	    }
	  }

    remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/power.off %s",vmid)
    stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.off")
	  time.Sleep(1 * time.Second)

    return stdout,nil

	} else {
		remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/power.off %s",vmid)
    stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.off")
		return stdout,nil
	}
}


func guestPowerGetState(c *Config, vmid string) string {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestPowerGetState]\n")

  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s", vmid)
  stdout, _   := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.getstate")

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

func guestGetIpAddress(c *Config, vmid string, guest_net_timeout int) string {
	esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
	log.Printf("[provider-esxi / guestGetIpAddress]\n")

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
	for uptime < guest_net_timeout {
		//  Primary method to get IP
		remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null |grep -A 5 'deviceConfigId = 4000' |tail -1|grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5]).){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'",vmid)
		stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get ip_address method 1")
		ip_address  = stdout
		if ip_address != "" {
			return ip_address
		}

		time.Sleep(3 * time.Second)

		//  Get uptime if above failed.
		remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null | grep 'uptimeSeconds ='|sed 's/^.*= //g'|sed s/,//g", vmid)
		stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get uptime")
		if err != nil {
			return ""
		}
		uptime, _   = strconv.Atoi(stdout)
	}

	//
	// Alternate method to get IP
	//
  remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null | grep 'uptimeSeconds ='|sed 's/^.*= //g'|sed s/,//g", vmid)
	stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get uptime")
	uptime, _   = strconv.Atoi(stdout)
	if uptime > 120 {
		remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null | grep -m 1 '^   ipAddress = ' | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5]).){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'",vmid)
    stdout, _   = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get ip_address method 2")
		ip_address2  = stdout
		if ip_address2 != "" {
			return ip_address2
		}
	}

	return ""
}

func validateVirtualDiskSlot(slot string) string {
	var result string
	
	// Split on comma.
  fields := strings.Split(slot + ":UnSet", ":")

  // if using simple format
  if fields[1] == "UnSet" {
    fields[1] = fields[0]
    fields[0] = "0"
  }

  field0i, _ := strconv.Atoi(fields[0])
  field1i, _ := strconv.Atoi(fields[1])
  result = "ok"

  if field0i < 0 || field0i > 3 {
    result = "scsi controller id out of range"
  }
  if field1i < 0 || field1i > 15 {
    result = "scsi id out of range"
  }
  if field0i == 0 && field1i == 0 {
    result = "scsi id used by boot disk"
  }
  if field1i == 7 {
    result = "scsi id 7 not allowed"
  }

  return result
}
