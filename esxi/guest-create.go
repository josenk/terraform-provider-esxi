package esxi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)


func guestCREATE(c *Config, guest_name string, disk_store string,
	 src_path string, resource_pool_name string, memsize string, numvcpus string,
	 virtual_networks [4][3]string ) (string, error) {

  esxiSSHinfo := SshConnectionStruct{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestCREATE]")

  var remote_cmd, vmid, stdout string
	var out bytes.Buffer

	dst_path := fmt.Sprintf("vi://%s:%s@%s%s", c.Esxi_username, c.Esxi_password, c.Esxi_hostname, resource_pool_name)

	ovf_cmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false " +
		"-dm=thin --name='%s' --overwrite -ds='%s' '%s' '%s'",guest_name, disk_store, src_path, dst_path)
	cmd := exec.Command("/bin/bash", "-c", ovf_cmd)

  log.Println("[provider-esxi] ovf_cmd: " + ovf_cmd )

	cmd.Stdout = &out
	err := cmd.Run()
	log.Printf("[provider-esxi] ovftool output: %q\n", out.String())
	if err != nil {
		log.Println("[provider-esxi] There was an ovftool Error: " + err.Error())
		return err.Error(), err
	}

  remote_cmd = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | sort -n | " +
		"grep \"[0-9] * %s .*%s\" | awk '{print $1}' | " +
		"tail -1", guest_name, guest_name)

  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get vmid")
	vmid = strings.TrimSpace(string(stdout))
	log.Printf("[provider-esxi] get_vmid_cmd: %s", vmid)
	if err != nil {
		log.Printf("[provider-esxi] Failed get vmid_cmd: %s", stdout)
		return "Failed get vmid", err
	}

	//
	//  make updates to vmx file
	//
  err = updateVmx_contents(c, vmid, true, memsize, numvcpus, virtual_networks)
	if err != nil {
		return "Failed to update vmx file on esxi host, see esxi console/logs for more details.", err
	}

	_, err = guestPowerOn(c, vmid)
	if err != nil {
		return "Failed to power on.", err
	}

  return vmid,err
}
