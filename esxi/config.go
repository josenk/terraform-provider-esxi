package esxi

import (
	//"fmt"
	//"log"

	//"github.com/hashicorp/terraform/helper/pathorcontents"
	//"github.com/hashicorp/terraform/terraform"
)

type Config struct {
	Esxi_hostname string
	Esxi_hostport string
	Esxi_username string
	Esxi_password string
}

func (c *Config) validateEsxiCreds() error {
  return nil
}
