package esxi

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"strings"
	"time"
)

// Connect to esxi host using ssh
func connectToHost(esxiSSHinfo SshConnectionStruct) (*ssh.Client, *ssh.Session, error) {

	sshConfig := &ssh.ClientConfig{
		User: esxiSSHinfo.user,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				// Reply password to all questions
				answers := make([]string, len(questions))
				for i, _ := range answers {
					answers[i] = esxiSSHinfo.pass
				}

				return answers, nil
			}),
		},
	}

	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	esxi_hostandport := fmt.Sprintf("%s:%s", esxiSSHinfo.host, esxiSSHinfo.port)

	attempt := 10
	for attempt > 0 {
		client, err := ssh.Dial("tcp", esxi_hostandport, sshConfig)
		if err != nil {
			log.Printf("[runRemoteSshCommand] Retry connection: %d\n", attempt)
			attempt -= 1
			time.Sleep(1 * time.Second)
		} else {

			session, err := client.NewSession()
			if err != nil {
				client.Close()
				return nil, nil, fmt.Errorf("Session Connection Error")
			}

			return client, session, nil

		}
	}
	return nil, nil, fmt.Errorf("Client Connection Error")
}

//  Run any remote ssh command on esxi server and return results.
func runRemoteSshCommand(esxiSSHinfo SshConnectionStruct, remoteSshCommand string, shortCmdDesc string) (string, error) {
	log.Println("[runRemoteSshCommand] :" + shortCmdDesc)

	client, session, err := connectToHost(esxiSSHinfo)
	if err != nil {
		log.Println("[runRemoteSshCommand] Failed err: " + err.Error())
		return "Failed to ssh to esxi host", err
	}

	stdout_raw, err := session.CombinedOutput(remoteSshCommand)
	stdout := strings.TrimSpace(string(stdout_raw))
	log.Printf("[runRemoteSshCommand] cmd:/%s/\n stdout:/%s/\nstderr:/%s/\n", remoteSshCommand, stdout, err)

	client.Close()
	return stdout, err
}
