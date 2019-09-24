package esxi

import (
	"fmt"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// Connect to esxi host using ssh
func connectToHost(esxiSSHinfo SshConnectionStruct, attempt int) (*ssh.Client, *ssh.Session, error) {

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

	//attempt := 10
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

	var attempt int

	if remoteSshCommand == "vmware --version" {
		attempt = 3
	} else {
		attempt = 10
	}
	client, session, err := connectToHost(esxiSSHinfo, attempt)
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

//  Function to scp file to esxi host.
func writeContentToRemoteFile(esxiSSHinfo SshConnectionStruct, content string, path string, shortCmdDesc string) (string, error) {
	log.Println("[writeContentToRemoteFile] :" + shortCmdDesc)

	f, _ := ioutil.TempFile("", "")
	fmt.Fprintln(f, content)
	f.Close()
	defer os.Remove(f.Name())

	client, session, err := connectToHost(esxiSSHinfo, 10)
	if err != nil {
		log.Println("[writeContentToRemoteFile] Failed err: " + err.Error())
		return "Failed to ssh to esxi host", err
	}

	err = scp.CopyPath(f.Name(), path, session)
	if err != nil {
		log.Println("[writeContentToRemoteFile] Failed err: " + err.Error())
		return "Failed to scp file to esxi host", err
	}

	client.Close()
	return content, err
}
