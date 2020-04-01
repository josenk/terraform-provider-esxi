package esxi

import (
	"fmt"
	"log"
)

func virtualSwitchCreate(c *Config, virtual_switch_name string) error {
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualSwitchCreate]")

	var cmd_result, remote_cmd string
	var err error

	remote_cmd = fmt.Sprintf("esxcfg-vswitch -a \"%s\"", virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "create virtual switch")

	if err != nil {
		return fmt.Errorf("Unable to create virtual switch: %w", err)
	}

	if cmd_result != "" {
		return fmt.Errorf("Unable to create virtual switch: %s", cmd_result)
	}

	remote_cmd = fmt.Sprintf("esxcfg-vswitch -c \"%s\"", virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate virtual switch exists")

	if err != nil {
		return fmt.Errorf("Unable to validate virtual switch: %w", err)
	}

	if cmd_result != "1" {
		return fmt.Errorf("Unable to validate virtual switch: %s", cmd_result)
	}

	return nil
}

func virtualSwitchRead(c *Config, virtual_switch_name string) (string, error) {

	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[virtualSwitchRead] Begin")

	var remote_cmd, cmd_result string
	var err error

	remote_cmd = fmt.Sprintf("esxcfg-vswitch -c \"%s\"", virtual_switch_name)
	cmd_result, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "validate virtual switch exists")

	if err != nil {
		return "", fmt.Errorf("Unable to validate virtual switch: %w", err)
	}

	if cmd_result != "1" {
		return "", fmt.Errorf("Unable to validate virtual switch: %s", cmd_result)
	}

	// var virtual_disk_disk_store, virtual_disk_dir, virtual_disk_name string
	// var virtual_disk_type, flatSize string
	// var virtual_disk_size int
	// var flatSizei64 int64
	// var s []string

	// //  Split virtdisk_id into it's variables
	// s = strings.Split(virtdisk_id, "/")
	// log.Printf("[virtualDiskREAD] len=%d cap=%d %v\n", len(s), cap(s), s)
	// if len(s) < 6 {
	// 	return "", "", "", 0, "", nil
	// }
	// virtual_disk_disk_store = s[3]
	// virtual_disk_dir = s[4]
	// virtual_disk_name = s[5]

	// // Test if virtual disk exists
	// remote_cmd := fmt.Sprintf("test -s \"%s\"", virtdisk_id)
	// _, err := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "test if virtual disk exists")
	// if err != nil {
	// 	return "", "", "", 0, "", err
	// }

	// //  Get virtual disk flat size
	// s = strings.Split(virtual_disk_name, ".")
	// if len(s) < 2 {
	// 	return "", "", "", 0, "", err
	// }
	// virtual_disk_nameFlat := fmt.Sprintf("%s-flat.%s", s[0], s[1])

	// remote_cmd = fmt.Sprintf("ls -l \"/vmfs/volumes/%s/%s/%s\" | awk '{print $5}'",
	// 	virtual_disk_disk_store, virtual_disk_dir, virtual_disk_nameFlat)
	// flatSize, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get size")
	// if err != nil {
	// 	return "", "", "", 0, "", err
	// }
	// flatSizei64, _ = strconv.ParseInt(flatSize, 10, 64)
	// virtual_disk_size = int(flatSizei64 / 1024 / 1024 / 1024)

	// // Determine virtual disk type  (only works if Guest is powered off)
	// remote_cmd = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS Z- LVID:' && echo true", virtdisk_id)
	// isZeroedThick, _ := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get disk type.  Is zeroedthick.")

	// remote_cmd = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS -- LVID:' && echo true", virtdisk_id)
	// isEagerZeroedThick, _ := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get disk type.  Is eagerzeroedthick.")

	// remote_cmd = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'NOMP -- :' && echo true", virtdisk_id)
	// isThin, _ := runRemoteSshCommand(esxiSSHinfo, remote_cmd, "Get disk type.  Is thin.")

	// if isThin == "true" {
	// 	virtual_disk_type = "thin"
	// } else if isZeroedThick == "true" {
	// 	virtual_disk_type = "zeroedthick"
	// } else if isEagerZeroedThick == "true" {
	// 	virtual_disk_type = "eagerzeroedthick"
	// } else {
	// 	virtual_disk_type = "Unknown"
	// }

	// Return results
	return virtual_switch_name, err
}
