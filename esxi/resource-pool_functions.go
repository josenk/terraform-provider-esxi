package esxi

import (
	"fmt"
	"log"
  "strings"
	"regexp"
)



//  Check if Pool exists (by name )and return it's Pool ID.
func getPoolID(c *Config, resource_pool_name string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getPoolID]")

	resource_pool_name = strings.TrimSpace(resource_pool_name)

	if resource_pool_name == "/" || resource_pool_name == "Resources" {
		return "ha-root-pool", nil
	}

	result := strings.Split(resource_pool_name, "/")
  resource_pool_name = result[len(result)-1]

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
  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getPoolNAME]")

	var ResourcePoolName, fullResourcePoolName string

	resource_pool_id = strings.TrimSpace(resource_pool_id)
	fullResourcePoolName = ""

	if resource_pool_id == "ha-root-pool" {
		return "/", nil
	}

  // Get full Resource Pool Path
	remote_cmd := fmt.Sprintf("grep -A1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep '<path>'", resource_pool_id)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get resource pool path")
	if err != nil {
		log.Printf("[provider-esxi / getPoolNAME] Failed get resource pool PATH: %s", stdout)
		return "", err
	}

	re := regexp.MustCompile(`[/<>\n]`)
  result := re.Split(stdout, -1)

  for i := range result {

		ResourcePoolName = ""
    if result[i] != "path" && result[i] != "host" && result[i] != "user" && result[i] != "" {

			r := strings.NewReplacer("name>","", "</name","")
		  remote_cmd := fmt.Sprintf("grep -B1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep -o name.*name", result[i])
		  stdout, _ := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get resource pool name")
			stdout = r.Replace(stdout)
			ResourcePoolName = strings.TrimSpace(stdout)

			if ResourcePoolName != "" {
			  if result[i] == resource_pool_id {
			    fullResourcePoolName = fullResourcePoolName + ResourcePoolName
			  } else {
			  	fullResourcePoolName = fullResourcePoolName + ResourcePoolName + "/"
			  }
			}
		}
  }

	return fullResourcePoolName, nil

}
