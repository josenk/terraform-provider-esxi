package esxi

import (
	//"fmt"
	//"log"
)

type Config struct {
	esxiHostName string
	esxiHostPort string
	esxiUserName string
	esxiPassword string
}

func (c *Config) validateEsxiCreds() error {
  return nil
}
