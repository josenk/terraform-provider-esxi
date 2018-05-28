package esxi

import (
	"fmt"
	"log"
	//"strconv"
	//"strings"
)

func GuestDelete(esxi_hostname string, esxi_hostport string, esxi_username string, esxi_password string,
  vmid string) int {
    encoded_esxi_password := esxi_password

    esxi_hostandport := fmt.Sprintf("%s:%s", esxi_hostname, esxi_hostport)
  	client, session, err := connectToHost(esxi_username, encoded_esxi_password, esxi_hostandport)
  	if err != nil {
  		log.Println("[provider-esxi] Failed to ssh to esxi host: " + err.Error())
  		return 1
    }


    remote_cmd := fmt.Sprintf("vim-cmd vmsvc/destroy %s", vmid)
  	stdout, err := session.CombinedOutput(remote_cmd)
  	//vmid := strings.TrimSpace(string(stdout))
  	log.Printf("[provider-esxi] guest destroy stdout: %s", stdout)
  	if err != nil {
  		log.Println("[provider-esxi] There was an Error destroying vmid: " + err.Error())
  		return 1
  	}



  	client.Close()
    return 0
  }
