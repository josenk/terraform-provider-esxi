package esxi

func getConnectionInfo(c *Config) ConnectionStruct {
	esxiConnInfo := ConnectionStruct{c.esxiHostName, c.esxiHostSSHport, c.esxiHostSSLport, c.esxiUserName, c.esxiPassword}

	return esxiConnInfo
}
