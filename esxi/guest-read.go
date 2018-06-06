package esxi

import (
	"fmt"
	"strings"
	"bufio"
	"regexp"
	"log"
)


func GuestRead(c *Config, vmid string) (string, string, string, error) {

  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  var guest_name, esxi_disk_store, esxi_resource_pool string
	r,_ := regexp.Compile("")

  short_desc := "Guest summary"
  remote_cmd := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, short_desc)

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
      esxi_disk_store = r.FindString(scanner.Text())
			nr := strings.NewReplacer("[","", "]","")
			esxi_disk_store = nr.Replace(esxi_disk_store)
    }
  }

  short_desc = "check if guest is in resource pool"
  remote_cmd = fmt.Sprintf(`grep -A2 'objID>%s</objID' /etc/vmware/hostd/pools.xml | grep -o resourcePool.*resourcePool`, vmid)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, short_desc)
  nr := strings.NewReplacer("resourcePool>","", "</resourcePool","")
  vm_resource_pool_id := nr.Replace(stdout)
	vm_resource_pool_id = strings.TrimSpace(vm_resource_pool_id)
	log.Printf("[provider-esxi / GuestRead] esxi_resource_pool|%s| scanner.Text():|%s|", vm_resource_pool_id, stdout)
	esxi_resource_pool, err = getPoolNAME(c, vm_resource_pool_id)
	log.Printf("[provider-esxi / GuestRead] esxi_resource_pool|%s| scanner.Text():|%s|", vm_resource_pool_id, err)


  return guest_name, esxi_disk_store, esxi_resource_pool, err
}
