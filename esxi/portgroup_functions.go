package esxi

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/jszwec/csvutil"
)

type portgroupSecurityPolicy struct {
	AllowForgedTransmits  bool `csv:"AllowForgedTransmits"`
	AllowMACAddressChange bool `csv:"AllowMACAddressChange"`
	AllowPromiscuous      bool `csv:"AllowPromiscuous"`
}

func portgroupRead(c *Config, name string) (string, int, error) {
	esxiConnInfo := getConnectionInfo(c)
	log.Println("[portgroupRead]")

	var vswitch string
	var vlan int
	var err error

	//  get portgroup info
	remote_cmd := fmt.Sprintf("esxcli network vswitch standard portgroup list |grep -m 1 \"^%s  \"", name)

	stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup list")
	if stdout == "" {
		return "", 0, fmt.Errorf("Failed to list portgroup: %s\n%s\n", stdout, err)
	}

	re, _ := regexp.Compile("(  .*  )  +[0-9]+  +[0-9]+$")
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vswitch = strings.Trim(re.FindStringSubmatch(stdout)[1], " ")
	} else {
		vswitch = ""
	}

	re, _ = regexp.Compile("  +([0-9]+)$")
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vlan, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		vlan = 0
	}

	return vswitch, vlan, nil
}

func portgroupSecurityPolicyRead(c *Config, name string) (*portgroupSecurityPolicy, error) {
	esxiConnInfo := getConnectionInfo(c)

	remote_cmd := fmt.Sprintf("esxcli --formatter=csv network vswitch standard portgroup policy security get -p \"%s\"", name)
	stdout, err := runRemoteSshCommand(esxiConnInfo, remote_cmd, "portgroup security policy")
	if stdout == "" {
		return nil, fmt.Errorf("Failed to get the portgroup security policy: %s\n%s\n", stdout, err)
	}

	var policies []portgroupSecurityPolicy
	if err = csvutil.Unmarshal([]byte(stdout), &policies); err != nil || len(policies) != 1 {
		return nil, fmt.Errorf("Failed to parse the portgroup security policy: %s\n%s\n", stdout, err)
	}

	return &policies[0], nil
}
