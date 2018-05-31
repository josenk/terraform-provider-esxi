package esxi

import (
	"fmt"
	"log"
)


func GuestDelete(esxi_hostname string, esxi_hostport string, esxi_username string, encoded_esxi_password string,
	 vmid string) int {

    esxiSSHinfo := SshConnectionInfo{esxi_hostname, esxi_hostport, esxi_username, encoded_esxi_password}

    remote_cmd := fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)

		stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get vmid")
		if err != nil {
			log.Printf("[provider-esxi] Failed destroy vmid: %s", stdout)
			return 1
		}
    return 0
  }
