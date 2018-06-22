package esxi

import (
	"fmt"
	"log"
  "strings"
  "bufio"
  "regexp"
  "strconv"
)


func resourcePoolREAD(c *Config, pool_id string) (int, bool, int, string, int, bool, int, string, string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Println("[provider-esxi / resourcePoolREAD] Begin" )
  var cpu_shares, mem_shares string
  var cpu_min, cpu_max, mem_min, mem_max, tmpvar int
  var cpu_min_expandable, mem_min_expandable bool

  remote_cmd := fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_get %s", pool_id)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "resource pool_config_get")

  if strings.Contains(stdout, "deleted") == true {
    log.Printf("[provider-esxi] Already deleted: %s", err)
    return 0, false, 0, "", 0, false, 0, "", "", err
  }
  if err != nil {
    log.Printf("[provider-esxi] Failed to get %s: %s", "resource pool_config_get", err)
    return 0, false, 0, "", 0, false, 0, "", "", err
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
        cpu_min_expandable,_ = strconv.ParseBool(r.FindString(scanner.Text()))
      } else {
        mem_min_expandable,_ = strconv.ParseBool(r.FindString(scanner.Text()))
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
    return 0, false, 0, "", 0, false, 0, "", "", err
  }

  log.Printf("[provider-esxi / resourcePoolREAD] |%s|%s|%s|%s|%s|%s|%s|%s|%s|",
     cpu_min, cpu_min_expandable, cpu_max, cpu_shares, mem_min, mem_min_expandable,
     mem_max, mem_shares, resource_pool_name)
  return cpu_min, cpu_min_expandable, cpu_max, cpu_shares, mem_min,
   mem_min_expandable, mem_max, mem_shares, resource_pool_name, err
}
