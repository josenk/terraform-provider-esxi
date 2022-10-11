package esxi

import (
	"fmt"
	"log"
)

type Config struct {
	esxiHostName    string
	esxiHostSSHport string
	esxiHostSSLport string
	esxiUserName    string
	esxiPassword    string

	esxiRemoteOvfToolPath string
}

func (c *Config) validateEsxiCreds() error {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[validateEsxiCreds]\n")

	var remote_cmd string
	var err error

	remote_cmd = fmt.Sprintf("vmware --version")
	_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "Connectivity test, get vmware version")
	if err != nil {
		return fmt.Errorf("Failed to connect to esxi host: %s\n", err)
	}

	runRemoteSshCommand(esxiConnInfo, "mkdir -p ~", "Create home directory if missing")

	if c.esxiRemoteOvfToolPath != "" {
		remote_cmd = fmt.Sprintf("%s --version", c.esxiRemoteOvfToolPath)
		vsn, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "Checking installation of ovftool on ESXi")
		if err != nil {
			return fmt.Errorf("Failed to invoke ovftool on ESXi host: %+v\n", err)
		}

		log.Printf("Found ovftool on ESXi host: %s\n", vsn)
	}

	return nil
}
