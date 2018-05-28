package esxi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	//"golang.org/x/crypto/ssh"
	//"strconv"
	"strings"
)

func GuestCreate(esxi_hostname string, esxi_hostport string, esxi_username string, encoded_esxi_password string,
	 guest_name string, esxi_disk_store string, src_path string, esxi_resource_pool string) (string, int) {

	dst_path := fmt.Sprintf("vi://%s:%s@%s/%s", esxi_username, encoded_esxi_password, esxi_hostname, esxi_resource_pool)

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

	esxi_hostandport := fmt.Sprintf("%s:%s", esxi_hostname, esxi_hostport)
	client, session, err := connectToHost(esxi_username, encoded_esxi_password, esxi_hostandport)
	if err != nil {
		log.Println("[provider-esxi] Failed to ssh to esxi host: " + err.Error())
		return "Failed to ssh to esxi host",1
  }



  remote_cmd := fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | sort -n | " +
		"grep \"[0-9] * %s .*%s\" | awk '{print $1}' | " +
		"tail -1", guest_name, guest_name)

	stdout, err := session.CombinedOutput(remote_cmd)
	vmid := strings.TrimSpace(string(stdout))
	log.Printf("[provider-esxi] get_vmid_cmd: %s", vmid)
	if err != nil {
		log.Println("[provider-esxi] There was an Error getting vmid: " + err.Error())
		return "There was an Error getting vmid",1
	}



	client.Close()
  return vmid, 0
}
