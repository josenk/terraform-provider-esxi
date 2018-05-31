package esxi

import (
	"golang.org/x/crypto/ssh"
	"fmt"
	"log"
)

// Connect to esxi host using ssh
//func connectToHost(user string, pass string, host string) (*ssh.Client, *ssh.Session, error) {
func connectToHost(esxiSSHinfo SshConnectionInfo) (*ssh.Client, *ssh.Session, error) {

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
func runRemoteSshCommand(esxiSSHinfo SshConnectionInfo, remoteSshCommand string, shortCmdDesc string) (string, error){

	client, session, err := connectToHost(esxiSSHinfo)
	if err != nil {
		log.Println("[provider-esxi] Failed to ssh to esxi host: " + err.Error())
		return "Failed to ssh to esxi host", err
  }

	stdout_raw, err := session.CombinedOutput(remoteSshCommand)
	stdout := string(stdout_raw)
	if err != nil {
		errorMessage := fmt.Sprintf("[provider-esxi] There was an ERROR to %s: %s",shortCmdDesc, err.Error())
		log.Println(errorMessage)
		return errorMessage, err
	}

	client.Close()
	return stdout, nil
}
