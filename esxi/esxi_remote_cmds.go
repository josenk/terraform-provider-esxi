package esxi

import (
	"golang.org/x/crypto/ssh"
	"fmt"
	"log"
)

// Connect to esxi host using ssh
func connectToHost(esxiSSHinfo SshConnectionStruct) (*ssh.Client, *ssh.Session, error) {

	sshConfig := &ssh.ClientConfig{
		User: esxiSSHinfo.user,
		Auth: []ssh.AuthMethod{ssh.Password(esxiSSHinfo.pass)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

  esxi_hostandport := fmt.Sprintf("%s:%s", esxiSSHinfo.host, esxiSSHinfo.port)
	client, err := ssh.Dial("tcp", esxi_hostandport, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}



//  Run any remote ssh command on esxi server and return results.
func runRemoteSshCommand(esxiSSHinfo SshConnectionStruct, remoteSshCommand string, shortCmdDesc string) (string, error){

  log.Println("[provider-esxi / runRemoteSshCommand] :" + shortCmdDesc )
	client, session, err := connectToHost(esxiSSHinfo)
	if err != nil {
		log.Println("[provider-esxi / runRemoteSshCommand] Failed err: " + err.Error())
		return "Failed to ssh to esxi host", err
  }

  log.Println("[provider-esxi / runRemoteSshCommand] remoteSshCommand: |" + remoteSshCommand + "|")
	stdout_raw, err := session.CombinedOutput(remoteSshCommand)
	stdout := string(stdout_raw)
	log.Printf("[provider-esxi / runRemoteSshCommand] cmd stdout|%s| err:|%s|", stdout, err)

	client.Close()
	return stdout, err
}
