package esxi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)


func GuestCreate(esxi_hostname string, esxi_hostport string, esxi_username string, encoded_esxi_password string,
	 guest_name string, esxi_disk_store string, src_path string, esxi_resource_pool string) (string, int) {

  esxiSSHinfo := SshConnectionInfo{esxi_hostname, esxi_hostport, esxi_username, encoded_esxi_password}

	dst_path := fmt.Sprintf("vi://%s:%s@%s/%s", esxiSSHinfo.user, esxiSSHinfo.pass, esxiSSHinfo.host, esxi_resource_pool)

	ovf_cmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false " +
		"-dm=thin --name='%s' --overwrite -ds='%s' '%s' '%s'",guest_name, esxi_disk_store, src_path, dst_path)


	cmd := exec.Command("/bin/bash", "-c", ovf_cmd)
	var out bytes.Buffer

  log.Println("[provider-esxi] ovf_cmd: " + ovf_cmd )

	cmd.Stdout = &out
	err := cmd.Run()
	log.Printf("[provider-esxi] ovftool output: %q\n", out.String())
	if err != nil {
		//log.Print(err)
		log.Println("[provider-esxi] There was an ovftool Error: " + err.Error())
		return "There was an ovftool Error",1
	}

  remote_cmd := fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | sort -n | " +
		"grep \"[0-9] * %s .*%s\" | awk '{print $1}' | " +
		"tail -1", guest_name, guest_name)

  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get vmid")
	vmid := strings.TrimSpace(string(stdout))
	log.Printf("[provider-esxi] get_vmid_cmd: %s", vmid)
	if err != nil {
		log.Printf("[provider-esxi] Failed get vmid_cmd: %s", stdout)
		return "Failed", 1
	}


  return vmid,0
}
