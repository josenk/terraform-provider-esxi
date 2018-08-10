package esxi

import (
	"fmt"
	"log"
  "strconv"
  "strings"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRESOURCEPOOLCreate(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceRESOURCEPOOLCreate]" )

	var remote_cmd string
  var pool_id, parent_pool string
	var err error

  resource_pool_name := d.Get("resource_pool_name").(string)
  cpu_min            := d.Get("cpu_min").(int)
  cpu_min_expandable := d.Get("cpu_min_expandable").(string)
  cpu_max            := d.Get("cpu_max").(int)
  cpu_shares         := strings.ToLower(d.Get("cpu_shares").(string))
  mem_min            := d.Get("mem_min").(int)
  mem_min_expandable := d.Get("mem_min_expandable").(string)
  mem_max            := d.Get("mem_max").(int)
  mem_shares         := strings.ToLower(d.Get("mem_shares").(string))
	parent_pool         = "Resources"

  if resource_pool_name == string('/') {
    return fmt.Errorf("Missing required resource_pool_name")
  }

  if resource_pool_name[0] == '/' {
    return fmt.Errorf("Resource Pool Name cannot start with /")
  }
  i := strings.LastIndex(resource_pool_name, "/")
  if i > 2 {
    parent_pool = resource_pool_name[:i]
    resource_pool_name = resource_pool_name[i+1:]
  }

  //  Check if already exists
	stdout, _ := getPoolID(c, resource_pool_name)
	if stdout != "" {
		d.SetId(stdout)
		return nil
	}

	cpu_min_opt := "--cpu-min=0"
	if cpu_min > 0 {
		cpu_min_opt = fmt.Sprintf("--cpu-min=%d", cpu_min)
	}

	cpu_min_expandable_opt := "--cpu-min-expandable=true"
	if cpu_min_expandable == "false" {
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

	mem_min_opt := "--mem-min=0"
	if mem_min > 0 {
		mem_min_opt = fmt.Sprintf("--mem-min=%d", mem_min)
	}

	mem_min_expandable_opt := "--mem-min-expandable=true"
	if mem_min_expandable == "false" {
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

	parent_pool_id, err := getPoolID(c, parent_pool)
	if err != nil {
		d.SetId("")
		return err
	}

	remote_cmd = fmt.Sprintf("vim-cmd hostsvc/rsrc/create %s %s %s %s %s %s %s %s %s %s",
		cpu_min_opt,cpu_min_expandable_opt, cpu_max_opt, cpu_shares_opt,
		mem_min_opt,mem_min_expandable_opt, mem_max_opt, mem_shares_opt, parent_pool_id, resource_pool_name)

	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create resource pool")
	pool_id, _  = getPoolID(c, resource_pool_name)
	if err != nil {
		d.SetId("")
		return err
	}

  //  Set pool_id
	d.SetId(pool_id)

	// Refresh
	resource_pool_name, cpu_min, cpu_min_expandable, cpu_max, cpu_shares, mem_min, mem_min_expandable, mem_max, mem_shares, err = resourcePoolRead(c, pool_id)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("resource_pool_name", resource_pool_name)
	d.Set("cpu_min", cpu_min)
	d.Set("cpu_min_expandable", cpu_min_expandable)
	d.Set("cpu_max", cpu_max)
	d.Set("cpu_shares", cpu_shares)
	d.Set("mem_min", mem_min)
	d.Set("mem_min_expandable", mem_min_expandable)
	d.Set("mem_max", mem_max)
	d.Set("mem_shares", mem_shares)

	return nil
}
