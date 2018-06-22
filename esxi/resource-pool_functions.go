package esxi

import (
	"fmt"
	"log"
  "strings"
)



//  Check if Pool exists (by name )and return it's Pool ID.
func getPoolID(c *Config, resource_pool_name string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getPoolID]")

	resource_pool_name = strings.TrimSpace(resource_pool_name)

  r := strings.NewReplacer("objID>","", "</objID","")
  remote_cmd := fmt.Sprintf("grep -A1 '<name>%s</name>' /etc/vmware/hostd/pools.xml | grep -o objID.*objID", resource_pool_name)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get existing resource pool id")
  if err == nil {
    stdout = r.Replace(stdout)
    stdout = strings.TrimSpace(stdout)
    return stdout, err
  } else {
    log.Printf("[provider-esxi / getPoolID] Failed get existing resource pool id: %s", stdout)
    return "", err
  }
}



//  Check if Pool exists (by id)and return it's Pool name.
func getPoolNAME(c *Config, resource_pool_id string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getPoolNAME]")
	
	resource_pool_id = strings.TrimSpace(resource_pool_id)

  r := strings.NewReplacer("name>","", "</name","")
  remote_cmd := fmt.Sprintf("grep -B1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep -o name.*name", resource_pool_id)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get existing resource pool name")
  if err == nil {
    stdout = r.Replace(stdout)
    stdout = strings.TrimSpace(stdout)
    return stdout, err
  } else {
    log.Printf("[provider-esxi / getPoolNAME] Failed get existing resource pool name: %s", stdout)
    return "", err
  }
}
