package esxi

import (
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"regexp"
	"log"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGUESTRead(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
	log.Println("[resourceGUESTRead]" )

  guest_startup_timeout  := d.Get("guest_startup_timeout").(int)

  var power string

  guest_name, disk_store, disk_size, boot_disk_type, resource_pool_name, memsize, numvcpus, virthwver, guestos, ip_address, virtual_networks, virtual_disks, power, err := guestREAD(c, d.Id(), guest_startup_timeout)
  if err != nil || guest_name == "" {
    d.SetId("")
    return nil
  }

  d.Set("guest_name",guest_name)
  d.Set("disk_store",disk_store)
  d.Set("disk_size",disk_size)
	if boot_disk_type != "Unknown" {
  	d.Set("boot_disk_type",boot_disk_type)
	}
  d.Set("resource_pool_name",resource_pool_name)
  d.Set("memsize",memsize)
  d.Set("numvcpus",numvcpus)
  d.Set("virthwver",virthwver)
  d.Set("guestos",guestos)
  d.Set("ip_address", ip_address)
  d.Set("power", power)

  // Do network interfaces
  log.Printf("virtual_networks: %q\n", virtual_networks)
  nics := make([]map[string]interface{}, 0, 1)

	for nic := 0; nic < 3; nic++ {
    if virtual_networks[nic][0] != "" {
		  out := make(map[string]interface{})
		  out["virtual_network"] = virtual_networks[nic][0]
      out["mac_address"]     = virtual_networks[nic][1]
      out["nic_type"]        = virtual_networks[nic][2]
		  nics = append(nics, out)
    }
	}
  d.Set("network_interfaces", nics)

	// Do virtual disks
  log.Printf("virtual_disks: %q\n", virtual_disks)
  vdisks := make([]map[string]interface{}, 0, 1)

	for vdisk := 0; vdisk < 3; vdisk++ {
    if virtual_disks[vdisk][0] != "" {
		  out := make(map[string]interface{})
		  out["virtual_disk_id"] = virtual_disks[vdisk][0]
      out["slot"]     = virtual_disks[vdisk][1]
		  vdisks = append(vdisks, out)
    }
	}
  d.Set("virtual_disks", vdisks)

  return nil
}


func guestREAD(c *Config, vmid string, guest_startup_timeout int) (string, string, string, string, string, string, string, string, string, string, [4][3]string, [60][2]string, string, error) {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[guestREAD]" )

  var guest_name, disk_store, virtual_disk_type, resource_pool_name, guestos, ip_address string
	var dst_vmx_ds, dst_vmx, dst_vmx_file, vmx_contents, power string
	var disk_size, vdiskindex int
	var memsize, numvcpus, virthwver string
	var virtual_networks [4][3]string
	var virtual_disks [60][2]string

	r,_ := regexp.Compile("")

  remote_cmd := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get Guest summary")

	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		return "", "", "", "", "", "", "", "", "", "", virtual_networks, virtual_disks, "", nil
	}

  scanner := bufio.NewScanner(strings.NewReader(stdout))
  for scanner.Scan() {
    switch {
    case strings.Contains(scanner.Text(),"name = "):
      r,_ = regexp.Compile(`\".*\"`)
      guest_name = r.FindString(scanner.Text())
			nr := strings.NewReplacer(`"`,"", `"`,"")
			guest_name = nr.Replace(guest_name)
    case strings.Contains(scanner.Text(),"vmPathName = "):
      r,_ = regexp.Compile(`\[.*\]`)
      disk_store = r.FindString(scanner.Text())
			nr := strings.NewReplacer("[","", "]","")
			disk_store = nr.Replace(disk_store)
    }
  }

  //  Get resource pool that this VM is located
  remote_cmd = fmt.Sprintf(`grep -A2 'objID>%s</objID' /etc/vmware/hostd/pools.xml | grep -o resourcePool.*resourcePool`, vmid)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "check if guest is in resource pool")
  nr := strings.NewReplacer("resourcePool>","", "</resourcePool","")
  vm_resource_pool_id := nr.Replace(stdout)
	log.Printf("[GuestRead] resource_pool_name|%s| scanner.Text():|%s|\n", vm_resource_pool_id, stdout)
	resource_pool_name, err = getPoolNAME(c, vm_resource_pool_id)
	log.Printf("[GuestRead] resource_pool_name|%s| scanner.Text():|%s|\n", vm_resource_pool_id, err)

	//
	//  Read vmx file into memory to read settings
	//
	//      -Get location of vmx file on esxi host
	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"",vmid)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx_ds")
	dst_vmx_ds  = stdout
	dst_vmx_ds  = strings.Trim(dst_vmx_ds, "[")
	dst_vmx_ds  = strings.Trim(dst_vmx_ds, "]")

	remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'",vmid)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx")
	dst_vmx     = stdout

	dst_vmx_file = "/vmfs/volumes/" + dst_vmx_ds + "/" + dst_vmx

	log.Printf("[guestREAD] dst_vmx_file: %s\n", dst_vmx_file)
	log.Printf("[guestREAD] disk_store: %s  dst_vmx_ds:%s\n", disk_store, dst_vmx_file)

  remote_cmd = fmt.Sprintf("cat \"%s\"", dst_vmx_file)
	vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "read guest_name.vmx file")

  // Used to keep track if a network interface is using static or generated macs.
  var isGeneratedMAC [3]bool

	//  Read vmx_contents line-by-line to get current settings.
	vdiskindex = 0
	scanner = bufio.NewScanner(strings.NewReader(vmx_contents))
  for scanner.Scan() {

    switch {
    case strings.Contains(scanner.Text(),"memSize = "):
      r,_ = regexp.Compile(`\".*\"`)
      stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			memsize = nr.Replace(stdout)
			log.Printf("[guestREAD] memsize found: %s\n", memsize)

    case strings.Contains(scanner.Text(),"numvcpus = "):
      r,_ = regexp.Compile(`\".*\"`)
      stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			numvcpus = nr.Replace(stdout)
			log.Printf("[guestREAD] numvcpus found: %s\n", numvcpus)

		case strings.Contains(scanner.Text(),"virtualHW.version = "):
      r,_ = regexp.Compile(`\".*\"`)
      stdout = r.FindString(scanner.Text())
			virthwver = strings.Replace(stdout,`"`,"",-1)
			log.Printf("[guestREAD] virthwver found: %s\n", virthwver)

		case strings.Contains(scanner.Text(),"guestOS = "):
      r,_ = regexp.Compile(`\".*\"`)
      stdout = r.FindString(scanner.Text())
			guestos = strings.Replace(stdout,`"`,"",-1)
			log.Printf("[guestREAD] guestos found: %s\n", guestos)

    case strings.Contains(scanner.Text(),"scsi"):
			re := regexp.MustCompile("scsi([0-3]):([0-9]{1,2}).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			if len(results) > 4 {
			  log.Printf("[guestREAD] %s : %s . %s = %s\n", results[1], results[2], results[3], results[4])

				if (results[1] == "0") && (results[2] == "0") {
					// Skip boot disk
				} else {
					if strings.Contains(results[3],"fileName") == true {
						log.Printf("[guestREAD] %s : %s\n", results[0], results[4])
						virtual_disks[vdiskindex][0] = results[4]
						virtual_disks[vdiskindex][1] = fmt.Sprintf("%s:%s", results[1], results[2])
						vdiskindex += 1
					}
				}
			}

		case strings.Contains(scanner.Text(),"ethernet"):
			re := regexp.MustCompile("ethernet(.).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			index,_ := strconv.Atoi(results[1])

			switch results[2] {
		  case "networkName":
				virtual_networks[index][0] = results[3]
				log.Printf("[guestREAD] %s : %s\n", results[0], results[3])

			case "addressType":
				if results[3] == "generated" {
					isGeneratedMAC[index] = true
				}

			case "generatedAddress":
				if isGeneratedMAC[index] == true {
					virtual_networks[index][1] = results[3]
					log.Printf("[guestREAD] %s : %s\n", results[0], results[3])
				}

			case "address":
				if isGeneratedMAC[index] == false {
					virtual_networks[index][1] = results[3]
					log.Printf("[resourceGUESTRead] %s : %s\n", results[0], results[3])
				}

			case "virtualDev":
					virtual_networks[index][2] = results[3]
					log.Printf("[guestREAD] %s : %s\n", results[0], results[3])
			}
    }

  }

	//  Get power state
	log.Println("guestREAD: guestPowerGetState")
	power = guestPowerGetState(c, vmid)

	//
	// Get IP address (need vmware tools installed)
	//
	if power == "on"  {
		ip_address = guestGetIpAddress(c, vmid, guest_startup_timeout)
		log.Printf("[guestREAD] guestGetIpAddress: %s\n", ip_address)
	} else {
		ip_address = ""
	}

	// Get boot disk size
	boot_disk_vmdkPATH,_ := getBootDiskPath(c, vmid)
	_, _, _, disk_size, virtual_disk_type, err = virtualDiskREAD(c, boot_disk_vmdkPATH)
	str_disk_size := strconv.Itoa(disk_size)

  // return results
  return guest_name, disk_store, str_disk_size, virtual_disk_type, resource_pool_name, memsize, numvcpus, virthwver, guestos, ip_address, virtual_networks, virtual_disks, power, err
}
