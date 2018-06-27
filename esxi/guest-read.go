package esxi

import (
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"regexp"
	"log"
)


func guestREAD(c *Config, vmid string) (string, string, string, string, string, [4][3]string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}

  var guest_name, disk_store, resource_pool_name string
	var dst_vmx_ds, dst_vmx, dst_vmx_file, vmx_contents string
	var memsize, numvcpus string
	var virtual_networks [4][3]string

	r,_ := regexp.Compile("")

  remote_cmd := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get Guest summary")

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


  remote_cmd = fmt.Sprintf(`grep -A2 'objID>%s</objID' /etc/vmware/hostd/pools.xml | grep -o resourcePool.*resourcePool`, vmid)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "check if guest is in resource pool")
  nr := strings.NewReplacer("resourcePool>","", "</resourcePool","")
  vm_resource_pool_id := nr.Replace(stdout)
	vm_resource_pool_id = strings.TrimSpace(vm_resource_pool_id)
	log.Printf("[provider-esxi / GuestRead] resource_pool_name|%s| scanner.Text():|%s|", vm_resource_pool_id, stdout)
	resource_pool_name, err = getPoolNAME(c, vm_resource_pool_id)
	log.Printf("[provider-esxi / GuestRead] resource_pool_name|%s| scanner.Text():|%s|", vm_resource_pool_id, err)

	//
	//  Read vmx file into memory to read settings
	//
	//      -Get location of vmx file on esxi host
	remote_cmd = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"",vmid)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx_ds")
	dst_vmx_ds  = strings.TrimSpace(string(stdout))
	dst_vmx_ds  = strings.Trim(dst_vmx_ds, "[")
	dst_vmx_ds  = strings.Trim(dst_vmx_ds, "]")

	remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'",vmid)
	stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx")
	dst_vmx     = strings.TrimSpace(string(stdout))

	dst_vmx_file = "/vmfs/volumes/" + dst_vmx_ds + "/" + dst_vmx

	log.Printf("[provider-esxi] dst_vmx_file: %s", dst_vmx_file)

  remote_cmd = fmt.Sprintf("cat %s", dst_vmx_file)
	vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "read guest_name.vmx file")

  // Used to keep track if a network interface is using static or generated macs.
  var isGeneratedMAC [3]bool

	//  Read vmx_contents line-by-line to get current settings.
	scanner = bufio.NewScanner(strings.NewReader(vmx_contents))
  for scanner.Scan() {

    switch {
    case strings.Contains(scanner.Text(),"memSize = "):
      r,_ = regexp.Compile(`\".*\"`)
      memsize = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			memsize = nr.Replace(memsize)
			log.Printf("[provider-esxi] memsize found: %s", memsize)

    case strings.Contains(scanner.Text(),"numvcpus = "):
      r,_ = regexp.Compile(`\".*\"`)
      numvcpus = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			numvcpus = nr.Replace(numvcpus)
			log.Printf("[provider-esxi] numvcpus found: %s", numvcpus)

		case strings.Contains(scanner.Text(),"ethernet"):
			re := regexp.MustCompile("ethernet(.).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			index,_ := strconv.Atoi(results[1])

			switch results[2] {
		  case "networkName":
				virtual_networks[index][0] = results[3]
				log.Printf("[provider-esxi] %s : %s", results[0], results[3])

			case "addressType":
				if results[3] == "generated" {
					isGeneratedMAC[index] = true
				}

			case "address":
				if isGeneratedMAC[index] == false {
					virtual_networks[index][1] = results[3]
					log.Printf("[provider-esxi] %s : %s", results[0], results[3])
				}

			case "virtualDev":
					virtual_networks[index][2] = results[3]
					log.Printf("[provider-esxi] %s : %s", results[0], results[3])
			}
    }
  }


  // return results
  return guest_name, disk_store, resource_pool_name, memsize, numvcpus, virtual_networks, err
}
