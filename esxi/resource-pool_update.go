package esxi

import (
	"fmt"
	"log"
  "strconv"
  "strings"
)


func resourcePoolUPDATE(c *Config, pool_id string, cpu_min int,
  cpu_min_expandable bool, cpu_max int, cpu_shares string, mem_min int,
  mem_min_expandable bool, mem_max int, mem_shares string) (string, error) {

  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
	log.Println("[provider-esxi / resourcePoolUPDATE] Begin" )
	
  var remote_cmd string

  cpu_min_opt := ""
  if cpu_min > 0 {
    cpu_min_opt = fmt.Sprintf("--cpu-min=%d", cpu_min)
  }

  cpu_min_expandable_opt := "--cpu-min-expandable=true"
  if cpu_min_expandable == false {
    cpu_min_expandable_opt = "--cpu-min-expandable=false"
  }

  cpu_max_opt := ""
  if cpu_max > 0 {
    cpu_max_opt = fmt.Sprintf("--cpu-max=%d", cpu_max)
  }

  cpu_shares_opt := "--cpu-shares=normal"
  if cpu_shares == "low" ||  cpu_shares == "high" {
    cpu_shares_opt = fmt.Sprintf("--cpu-shares=%s", cpu_shares)
  } else {
    tmp_var, err := strconv.Atoi(cpu_shares)
    if err == nil {
      cpu_shares_opt = fmt.Sprintf("--cpu-shares=%d", tmp_var)
    }
  }

  mem_min_opt := ""
  if mem_min > 0 {
    mem_min_opt = fmt.Sprintf("--mem-min=%d", mem_min)
  }

  mem_min_expandable_opt := "--mem-min-expandable=true"
  if mem_min_expandable == false {
    mem_min_expandable_opt = "--mem-min-expandable=false"
  }

  mem_max_opt := ""
  if mem_max > 0 {
    mem_max_opt = fmt.Sprintf("--mem-max=%d", mem_max)
  }

  mem_shares_opt := "--mem-shares=normal"
  if mem_shares == "low" ||  mem_shares == "high" {
    mem_shares_opt = fmt.Sprintf("--mem-shares=%s", mem_shares)
  } else {
    tmp_var, err := strconv.Atoi(mem_shares)
    if err == nil {
      mem_shares_opt = fmt.Sprintf("--mem-shares=%d", tmp_var)
    }
  }


  _, err := getPoolNAME(c, pool_id)
  if err != nil {
    return "", err
  }


  remote_cmd = fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_set %s %s %s %s %s %s %s %s %s",
    cpu_min_opt,cpu_min_expandable_opt, cpu_max_opt, cpu_shares_opt,
    mem_min_opt,mem_min_expandable_opt, mem_max_opt, mem_shares_opt, pool_id)

	stdout,_ := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create resource pool")
  log.Printf("[provider-esxi / resourcePoolUPDATE] stdout |%s|\n", stdout)

  r := strings.NewReplacer("'vim.ResourcePool:","", "'","")
  stdout = r.Replace(stdout)
  return stdout,err
}
