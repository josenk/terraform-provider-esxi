package esxi

import (
	"fmt"
	"log"
	"regexp"
)

type Config struct {
	esxiHostName    string
	esxiHostSSHport string
	esxiHostSSLport string
	esxiUserName    string
	esxiPassword    string
	esxiVersion     string
}

func (c *Config) validateEsxiCreds() error {
	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[validateEsxiCreds]\n")

	var remote_cmd, raw_esxi_version string
	var err error

	remote_cmd = fmt.Sprintf("vmware --version")
	raw_esxi_version, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "Connectivity test, get vmware version")
	if err != nil {
		return fmt.Errorf("Failed to connect to esxi host: %s\n", err)
	}

	re := regexp.MustCompile(`(\d{1,2}\.\d{1,2}\.\d{1,3})`)
	c.esxiVersion = re.FindAllString(raw_esxi_version, -1)[0]

	runRemoteSshCommand(esxiConnInfo, "mkdir -p ~", "Create home directory if missing")

	return nil
}
