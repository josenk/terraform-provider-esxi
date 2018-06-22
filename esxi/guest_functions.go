package esxi

import (
	"fmt"
	"strings"
	"log"
  "regexp"
)



func getDst_vmx_file(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getDst_vmx_file]")

  //      -Get location of vmx file on esxi host
  var dst_vmx_ds, dst_vmx, dst_vmx_file string
  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"",vmid)
	stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx_ds")
	dst_vmx_ds   = strings.TrimSpace(string(stdout))
	dst_vmx_ds   = strings.Trim(dst_vmx_ds, "[")
	dst_vmx_ds   = strings.Trim(dst_vmx_ds, "]")

	remote_cmd   = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'",vmid)
	stdout, err  = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "get dst_vmx")
	dst_vmx      = strings.TrimSpace(string(stdout))

	dst_vmx_file = "/vmfs/volumes/" + dst_vmx_ds + "/" + dst_vmx
  return dst_vmx_file, err
}

func readVmx_contents(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / getVmx_contents]")

  var remote_cmd, vmx_contents string

  dst_vmx_file,err := getDst_vmx_file(c, vmid)
  remote_cmd = fmt.Sprintf("cat %s", dst_vmx_file)
  vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "read guest_name.vmx file")

  return vmx_contents, err
}


func updateVmx_contents(c *Config, vmid string, memsize string, numvcpus string) error {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / updateVmx_contents]")
  var regexReplacement, remote_cmd string

  vmx_contents, err := readVmx_contents(c, vmid)
	if err != nil {
		log.Printf("[provider-esxi] Failed get vmx contents: %s", err)
		return err
	}

	// modify memsize
  if memsize != "" {
		re := regexp.MustCompile("memSize = \".*\"")
		regexReplacement = fmt.Sprintf("memSize = \"%s\"", memsize)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}

	// modify numvcpus
	if numvcpus != "" {
		re := regexp.MustCompile("numvcpus = \".*\"")
		regexReplacement = fmt.Sprintf("numvcpus = \"%s\"", numvcpus)
		vmx_contents = re.ReplaceAllString(vmx_contents, regexReplacement)
	}



	//
	//  Write vmx file to esxi host
	//
	vmx_contents = strings.Replace(vmx_contents, "\"", "\\\"", -1)
	log.Printf("[provider-esxi] New guest_name.vmx: %s", vmx_contents)

  dst_vmx_file,err := getDst_vmx_file(c, vmid)
  remote_cmd = fmt.Sprintf("echo \"%s\" >%s", vmx_contents, dst_vmx_file)
	vmx_contents, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "write guest_name.vmx file")

	remote_cmd  = fmt.Sprintf("vim-cmd vmsvc/reload %s",vmid)
	_, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/reload")
  return err
}

//func guestPowerGetState(c *Config, vmid string) (string, error) {
//  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
//  log.Printf("[provider-esxi / guestPowerGetState]")
//
//  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s",vmid)
//  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.on")
//  return stdout,err
//}

func guestPowerOn(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestPowerOn]")

  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.on %s",vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.on")
  return stdout,err
}

func guestPowerOff(c *Config, vmid string) (string, error) {
  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
  log.Printf("[provider-esxi / guestPowerOff]")

  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.off %s",vmid)
  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.off")
  return stdout,err
}

//func guestShutdown(c *Config, vmid string) (string, error) {
//  esxiSSHinfo := SshConnectionInfo{c.Esxi_hostname, c.Esxi_hostport, c.Esxi_username, c.Esxi_password}
//  log.Printf("[provider-esxi / guestShutdown]")
//
//  remote_cmd  := fmt.Sprintf("vim-cmd vmsvc/power.off %s",vmid)
//  stdout, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "vmsvc/power.shutdown")
//  return stdout, err
//}
