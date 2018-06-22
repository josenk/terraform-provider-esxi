package esxi

import (
	"fmt"
	"strings"
	"bufio"
	"regexp"
	"log"
)


func guestREAD(c *Config, vmid string) (string, string, string, string, string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}

  var guest_name, disk_store, resource_pool_name string
	var dst_vmx_ds, dst_vmx, dst_vmx_file, vmx_contents string
	var memsize, numvcpus string
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


	scanner = bufio.NewScanner(strings.NewReader(vmx_contents))
  for scanner.Scan() {
    switch {
    case strings.Contains(scanner.Text(),"memSize = "):
      r,_ = regexp.Compile(`\".*\"`)
      memsize = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			memsize = nr.Replace(memsize)
    case strings.Contains(scanner.Text(),"numvcpus = "):
      r,_ = regexp.Compile(`\".*\"`)
      numvcpus = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`,"", `"`,"")
			numvcpus = nr.Replace(numvcpus)

    }
  }


  return guest_name, disk_store, resource_pool_name, memsize, numvcpus, err
}
