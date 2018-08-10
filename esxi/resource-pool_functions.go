package esxi

import (
	"fmt"
	"log"
	"bufio"
  "strings"
	"strconv"
	"regexp"
	"errors"
)


//  Check if Pool exists (by name )and return it's Pool ID.
func getPoolID(c *Config, resource_pool_name string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
  log.Printf("[getPoolID]\n")

	if resource_pool_name == "/" || resource_pool_name == "Resources" {
		return "ha-root-pool", nil
	}

	result := strings.Split(resource_pool_name, "/")
  resource_pool_name = result[len(result)-1]

  r := strings.NewReplacer("objID>","", "</objID","")
  remote_cmd := fmt.Sprintf("grep -A1 '<name>%s</name>' /etc/vmware/hostd/pools.xml | grep -o objID.*objID | tail -1", resource_pool_name)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get existing resource pool id")
  if err == nil {
    stdout = r.Replace(stdout)
    return stdout, err
  } else {
    log.Printf("[getPoolID] Failed get existing resource pool id: %s\n", stdout)
    return "", err
  }
}



//  Check if Pool exists (by id)and return it's Pool name.
func getPoolNAME(c *Config, resource_pool_id string) (string, error) {
  esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
  log.Printf("[getPoolNAME]\n")

	var ResourcePoolName, fullResourcePoolName string

	fullResourcePoolName = ""

	if resource_pool_id == "ha-root-pool" {
		return "/", nil
	}

  // Get full Resource Pool Path
	remote_cmd := fmt.Sprintf("grep -A1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep '<path>'", resource_pool_id)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get resource pool path")
	if err != nil {
		log.Printf("[getPoolNAME] Failed get resource pool PATH: %s\n", stdout)
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
			ResourcePoolName = r.Replace(stdout)

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


func resourcePoolRead(c *Config, pool_id string) (string, int, string, int, string, int, string, int, string, error) {
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourcePoolRead]" )

	var remote_cmd, stdout, cpu_shares, mem_shares string
  var cpu_min, cpu_max, mem_min, mem_max, tmpvar int
  var cpu_min_expandable, mem_min_expandable string
	var err error


	remote_cmd = fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_get %s", pool_id)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "resource pool_config_get")

  if strings.Contains(stdout, "deleted") == true {
    log.Printf("[provider-esxi] Already deleted: %s\n", err)
		return "", 0, "", 0, "", 0, "", 0, "", nil
  }
  if err != nil {
    log.Printf("[provider-esxi] Failed to get %s: %s\n", "resource pool_config_get", err)
		return "", 0, "", 0, "", 0, "", 0, "", errors.New("Failed to get Resource Pool config.")
  }

  is_cpu_flag := true

  scanner := bufio.NewScanner(strings.NewReader(stdout))
  for scanner.Scan() {
    switch {
    case strings.Contains(scanner.Text(),"memoryAllocation = "):
      is_cpu_flag = false

    case strings.Contains(scanner.Text(),"reservation = "):
      r,_ := regexp.Compile("[0-9]+")
      if is_cpu_flag == true {
        cpu_min,_ = strconv.Atoi(r.FindString(scanner.Text()))
      } else {
        mem_min,_ = strconv.Atoi(r.FindString(scanner.Text()))
      }

    case strings.Contains(scanner.Text(),"expandableReservation = "):
      r,_ := regexp.Compile("(true|false)")
      if is_cpu_flag == true {
        cpu_min_expandable = r.FindString(scanner.Text())
      } else {
        mem_min_expandable = r.FindString(scanner.Text())
      }

    case strings.Contains(scanner.Text(),"limit = "):
      r,_ := regexp.Compile("-?[0-9]+")
			tmpvar,_ = strconv.Atoi(r.FindString(scanner.Text()))
			if tmpvar < 0 {
				tmpvar = 0
			}
      if is_cpu_flag == true {
        cpu_max = tmpvar
      } else {
        mem_max = tmpvar
      }

    case strings.Contains(scanner.Text(),"shares = "):
      r,_ := regexp.Compile("[0-9]+")
      if is_cpu_flag == true {
        cpu_shares = r.FindString(scanner.Text())
      } else {
        mem_shares = r.FindString(scanner.Text())
      }

    case strings.Contains(scanner.Text(),"level = "):
      r,_ := regexp.Compile("(low|high|normal)")
      if r.FindString(scanner.Text()) != "" {
        if is_cpu_flag == true {
          cpu_shares = r.FindString(scanner.Text())
        } else {
          mem_shares = r.FindString(scanner.Text())
        }
      }
    }
  }

  resource_pool_name, err := getPoolNAME(c, pool_id)
  if err != nil {
    log.Printf("[provider-esxi] Failed to get Resource Pool name: %s\n", err)
		return "", 0, "", 0, "", 0, "", 0, "", errors.New("Failed to get Resource Pool name.")
  }

  return  resource_pool_name, cpu_min, cpu_min_expandable, cpu_max, cpu_shares,
	  mem_min, mem_min_expandable, mem_max, mem_shares, nil
}
