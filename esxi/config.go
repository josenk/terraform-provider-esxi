package esxi

import (
	//"fmt"
	//"log"
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
