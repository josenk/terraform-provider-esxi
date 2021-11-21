package esxi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func guestCREATE(c *Config, guest_name string, disk_store string,
	src_path string, resource_pool_name string, strmemsize string, strnumvcpus string, strvirthwver string, guestos string,
	boot_disk_type string, boot_disk_size string, virtual_networks [10][4]string,
	virtual_disks [60][2]string, guest_shutdown_timeout int, ovf_properties_timer int, notes string,
	guestinfo map[string]interface{}, ovf_properties map[string]string) (string, error) {

	esxiConnInfo := getConnectionInfo(c)
	log.Printf("[guestCREATE]\n")

	var memsize, numvcpus, virthwver int
	var boot_disk_vmdkPATH, remote_cmd, vmid, stdout, vmx_contents string
	var osShellCmd, osShellCmdOpt string
	var out bytes.Buffer
	var err error
	var is_ovf_properties bool
	var ovf_bat *os.File
	err = nil
	is_ovf_properties = false

	memsize, _ = strconv.Atoi(strmemsize)
	numvcpus, _ = strconv.Atoi(strnumvcpus)
	virthwver, _ = strconv.Atoi(strvirthwver)

	//
	//  Check if Disk Store already exists
	//
	err = diskStoreValidate(c, disk_store)
	if err != nil {
		return "", fmt.Errorf("Failed to validate disk store: %s\n", err)
	}

	//
	//  Check if guest already exists
	//
	// get VMID (by name)
	vmid, err = guestGetVMID(c, guest_name)

	if vmid != "" {
		// We don't need to create the VM.   It already exists.
		fmt.Printf("[guestCREATE] guest %s already exists vmid: %s\n", guest_name, stdout)

		//
		//   Power off guest if it's powered on.
		//
		currentpowerstate := guestPowerGetState(c, vmid)
		if currentpowerstate == "on" || currentpowerstate == "suspended" {
			_, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
			if err != nil {
				return "", fmt.Errorf("Failed to power off: %s\n", err)
			}
		}

	} else if src_path == "none" {

		// check if path already exists.
		fullPATH := fmt.Sprintf("/vmfs/volumes/%s/%s", disk_store, guest_name)
		boot_disk_vmdkPATH = fmt.Sprintf("\"/vmfs/volumes/%s/%s/%s.vmdk\"", disk_store, guest_name, guest_name)

		remote_cmd = fmt.Sprintf("ls -d %s", boot_disk_vmdkPATH)
		stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "check if guest path already exists.")
		if strings.Contains(stdout, "No such file or directory") != true {
			fmt.Printf("Error: Guest may already exists. vmdkPATH:%s\n", boot_disk_vmdkPATH)
			return "", fmt.Errorf("Guest may already exists. vmdkPATH:%s\n", boot_disk_vmdkPATH)
		}

		remote_cmd = fmt.Sprintf("ls -d \"%s\"", fullPATH)
		stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "check if guest path already exists.")
		if strings.Contains(stdout, "No such file or directory") == true {
			remote_cmd = fmt.Sprintf("mkdir \"%s\"", fullPATH)
			stdout, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "create guest path")
			if err != nil {
				log.Printf("[guestCREATE] Failed to create guest path. fullPATH:%s\n", fullPATH)
				return "", fmt.Errorf("Failed to create guest path. fullPATH:%s\n", fullPATH)
			}
		}

		hasISO := false
		isofilename := ""
		notes = strings.Replace(notes, "\"", "|22", -1)

		if numvcpus == 0 {
			numvcpus = 1
		}
		if memsize == 0 {
			memsize = 512
		}
		if virthwver == 0 {
			virthwver = 13
		}
		if guestos == "" {
			guestos = "centos-64"
		}
		if boot_disk_size == "" {
			boot_disk_size = "16"
		}

		// Build VM by default/black config
		vmx_contents =
			fmt.Sprintf("config.version = \\\"8\\\"\n") +
				fmt.Sprintf("virtualHW.version = \\\"%d\\\"\n", virthwver) +
				fmt.Sprintf("displayName = \\\"%s\\\"\n", guest_name) +
				fmt.Sprintf("numvcpus = \\\"%d\\\"\n", numvcpus) +
				fmt.Sprintf("memSize = \\\"%d\\\"\n", memsize) +
				fmt.Sprintf("guestOS = \\\"%s\\\"\n", guestos) +
				fmt.Sprintf("annotation = \\\"%s\\\"\n", notes) +
				fmt.Sprintf("floppy0.present = \\\"FALSE\\\"\n") +
				fmt.Sprintf("scsi0.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("scsi0.sharedBus = \\\"none\\\"\n") +
				fmt.Sprintf("scsi0.virtualDev = \\\"lsilogic\\\"\n") +
				fmt.Sprintf("disk.EnableUUID = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge0.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge4.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge4.virtualDev = \\\"pcieRootPort\\\"\n") +
				fmt.Sprintf("pciBridge4.functions = \\\"8\\\"\n") +
				fmt.Sprintf("pciBridge5.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge5.virtualDev = \\\"pcieRootPort\\\"\n") +
				fmt.Sprintf("pciBridge5.functions = \\\"8\\\"\n") +
				fmt.Sprintf("pciBridge6.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge6.virtualDev = \\\"pcieRootPort\\\"\n") +
				fmt.Sprintf("pciBridge6.functions = \\\"8\\\"\n") +
				fmt.Sprintf("pciBridge7.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("pciBridge7.virtualDev = \\\"pcieRootPort\\\"\n") +
				fmt.Sprintf("pciBridge7.functions = \\\"8\\\"\n") +
				fmt.Sprintf("scsi0:0.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("scsi0:0.fileName = \\\"%s.vmdk\\\"\n", guest_name) +
				fmt.Sprintf("scsi0:0.deviceType = \\\"scsi-hardDisk\\\"\n")
		if hasISO == true {
			vmx_contents = vmx_contents +
				fmt.Sprintf("ide1:0.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("ide1:0.fileName = \\\"emptyBackingString\\\"\n") +
				fmt.Sprintf("ide1:0.deviceType = \\\"atapi-cdrom\\\"\n") +
				fmt.Sprintf("ide1:0.startConnected = \\\"FALSE\\\"\n") +
				fmt.Sprintf("ide1:0.clientDevice = \\\"TRUE\\\"\n")
		} else {
			vmx_contents = vmx_contents +
				fmt.Sprintf("ide1:0.present = \\\"TRUE\\\"\n") +
				fmt.Sprintf("ide1:0.fileName = \\\"%s\\\"\n", isofilename) +
				fmt.Sprintf("ide1:0.deviceType = \\\"cdrom-image\\\"\n")
		}

		//
		//  Write vmx file to esxi host
		//
		log.Printf("[guestCREATE] New guest_name.vmx: %s\n", vmx_contents)

		dst_vmx_file := fmt.Sprintf("%s/%s.vmx", fullPATH, guest_name)

		remote_cmd = fmt.Sprintf("echo \"%s\" >\"%s\"", vmx_contents, dst_vmx_file)
		vmx_contents, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "write guest_name.vmx file")

		//  Create boot disk (vmdk)
		remote_cmd = fmt.Sprintf("vmkfstools -c %sG -d %s \"%s/%s.vmdk\"", boot_disk_size, boot_disk_type, fullPATH, guest_name)
		_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "vmkfstools (make boot disk)")
		if err != nil {
			remote_cmd = fmt.Sprintf("rm -fr \"%s\"", fullPATH)
			stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "cleanup guest path because of failed events")
			log.Printf("[guestCREATE] Failed to vmkfstools (make boot disk):%s\n", err.Error())
			return "", fmt.Errorf("Failed to vmkfstools (make boot disk):%s\n", err.Error())
		}

		poolID, err := getPoolID(c, resource_pool_name)
		log.Println("[guestCREATE] DEBUG: " + poolID)
		if err != nil {
			log.Printf("[guestCREATE] Failed to use Resource Pool ID:%s\n", poolID)
			return "", fmt.Errorf("Failed to use Resource Pool ID:%s\n", poolID)
		}
		remote_cmd = fmt.Sprintf("vim-cmd solo/registervm \"%s\" %s %s", dst_vmx_file, guest_name, poolID)
		_, err = runRemoteSshCommand(esxiConnInfo, remote_cmd, "solo/registervm")
		if err != nil {
			log.Printf("[guestCREATE] Failed to register guest:%s\n", err.Error())
			remote_cmd = fmt.Sprintf("rm -fr \"%s\"", fullPATH)
			stdout, _ = runRemoteSshCommand(esxiConnInfo, remote_cmd, "cleanup guest path because of failed events")
			return "", fmt.Errorf("Failed to register guest:%s\n", err.Error())
		}

	} else {
		//  Build VM by ovftool

		//  Check if source file exist.
		if strings.HasPrefix(src_path, "http://") || strings.HasPrefix(src_path, "https://") {
			log.Printf("[guestCREATE] Source is URL.\n")
			resp, err := http.Get(src_path)
			if (err != nil) || (resp.StatusCode != 200) {
				log.Printf("[guestCREATE] URL not accessible: %s\n", src_path)
				log.Printf("[guestCREATE] URL StatusCode: %d\n", resp.StatusCode)
				log.Printf("[guestCREATE] URL Error: %s\n", err.Error())
				defer resp.Body.Close()
				return "", fmt.Errorf("URL not accessible: %s\n%s", src_path, err.Error())
			}
			defer resp.Body.Close()
		} else if strings.HasPrefix(src_path, "vi://") {
			log.Printf("[guestCREATE] Source is Guest VM (vi).\n")
		} else {
			log.Printf("[guestCREATE] Source is local.\n")
			if _, err := os.Stat(src_path); os.IsNotExist(err) {
				log.Printf("[guestCREATE] File not found, Error: %s\n", err.Error())
				return "", fmt.Errorf("File not found: %s\n", src_path)
			}
		}

		//  Set params for ovftool
		if boot_disk_type == "zeroedthick" {
			boot_disk_type = "thick"
		}

		username := url.QueryEscape(c.esxiUserName)
		password := url.QueryEscape(c.esxiPassword)
		dst_path := fmt.Sprintf("vi://%s:%s@%s:%s/%s", username, password, c.esxiHostName, c.esxiHostSSLport, resource_pool_name)

		net_param := ""
		if (strings.HasSuffix(src_path, ".ova") || strings.HasSuffix(src_path, ".ovf")) && virtual_networks[0][0] != "" {
			// if ovf_network is set (not emptied), then we will add --net to net_param for ovf network mapping
			// if ovt_network is emptied, then we will just resume to the default --network creation
			log.Printf("[guestCREATE] Detecting single --network or multiple --net...\n")
			if virtual_networks[0][3] == "" {
				net_param += " --network='" + virtual_networks[0][0] + "'"
			}

			for i := 0; i < 10; i++ {
				if virtual_networks[i][3] != "" {
					net_param += " --net:'" + virtual_networks[i][3] + "=" + virtual_networks[i][0] + "'"
				}
			}
			log.Printf("[guestCREATE] net_param: %s\n", net_param)
		}

		extra_params := ""
		if (len(ovf_properties) > 0) && (strings.HasSuffix(src_path, ".ova") || strings.HasSuffix(src_path, ".ovf")) {
			is_ovf_properties = true
			// in order to process any OVF params, guest should be immediately powered on
			// This is because the ESXi host doesn't have a cache to store the OVF parameters, like the vCenter Server does.
			// Therefore, you MUST use the ‘--X:injectOvfEnv’ option with the ‘--poweron’ option
			extra_params = "--X:injectOvfEnv --allowExtraConfig --powerOn "

			for ovf_prop_key, ovf_prop_value := range ovf_properties {
				extra_params = fmt.Sprintf("%s --prop:%s='%s' ", extra_params, ovf_prop_key, ovf_prop_value)
			}
			log.Println("[guestCREATE] ovf_properties extra_params: " + extra_params)
		}

		ovf_cmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false %s "+
			"-dm=%s --name='%s' --overwrite -ds='%s' %s '%s' '%s'", extra_params, boot_disk_type, guest_name, disk_store, net_param, src_path, dst_path)

		if runtime.GOOS == "windows" {
			osShellCmd = "cmd.exe"
			osShellCmdOpt = "/c"

			ovf_cmd = strings.Replace(ovf_cmd, "'", "\"", -1)

			ovf_bat, _ = ioutil.TempFile("", "ovf_cmd*.bat")

			_, err = os.Stat(ovf_bat.Name())

			// delete file if exists
			if os.IsExist(err) {
				err = os.Remove(ovf_bat.Name())
				if err != nil {
					return "", fmt.Errorf("Unable to delete existing %s: %s\n", ovf_bat.Name(), err.Error())
				}
			}

			//  create new batch file
			file, err := os.Create(ovf_bat.Name())
			if err != nil {
				defer file.Close()
				return "", fmt.Errorf("Unable to create %s: %s\n", ovf_bat.Name(), err.Error())
			}

			_, err = file.WriteString(strings.Replace(ovf_cmd, "%", "%%", -1))
			if err != nil {
				defer file.Close()
				return "", fmt.Errorf("Unable to write to %s: %s\n", ovf_bat.Name(), err.Error())
			}

			err = file.Close()
			if err != nil {
				defer file.Close()
				return "", fmt.Errorf("Unable to close %s: %s\n", ovf_bat.Name(), err.Error())
			}
			ovf_cmd = ovf_bat.Name()

		} else {
			osShellCmd = "/bin/bash"
			osShellCmdOpt = "-c"
		}

		//  Execute ovftool script (or batch) here.
		cmd := exec.Command(osShellCmd, osShellCmdOpt, ovf_cmd)

		re := regexp.MustCompile(`vi://.*?@`)
		log.Printf("[guestCREATE] ovf_cmd: %s\n", re.ReplaceAllString(ovf_cmd, "vi://XXXX:YYYY@"))

		cmd.Stdout = &out
		err = cmd.Run()
		log.Printf("[guestCREATE] ovftool output: %q\n", out.String())

		//  Attempt to delete tmp batch file.
		if ovf_bat != nil {
			_ = cmd.Wait()
			_ = os.Remove(ovf_bat.Name())
		}

		if err != nil {
			log.Printf("[guestCREATE] Failed, There was an ovftool Error: %s\n%s\n", out.String(), err.Error())
			return "", fmt.Errorf("There was an ovftool Error: %s\n%s\n", out.String(), err.Error())
		}
	}

	// get VMID (by name)
	vmid, err = guestGetVMID(c, guest_name)
	if err != nil {
		return "", fmt.Errorf("Failed to get vmid: %s\n", err)
	}

	//
	//   ovf_properties require ovftool to power on the VM to inject the properties.
	//   Unfortunatly, there is no way to know when cloud-init is finished?!?!?  Just need
	//   to wait for ovf_properties_timer seconds, then shutdown/power-off to continue...
	//
	if is_ovf_properties == true {
		currentpowerstate := guestPowerGetState(c, vmid)
		log.Printf("[guestCREATE] Current VM PowerState: %s\n", currentpowerstate)
		if currentpowerstate != "on" {
			return vmid, fmt.Errorf("[guestCREATE] Failed to poweron after ovf_properties injection.\n")
		}
		// allow cloud-init to process.
		duration := time.Duration(ovf_properties_timer) * time.Second

		log.Printf("[guestCREATE] Waiting for ovf_properties_timer: %s\n", duration)

		time.Sleep(duration)
		_, err = guestPowerOff(c, vmid, guest_shutdown_timeout)
		if err != nil {
			return vmid, fmt.Errorf("[guestCREATE] Failed to shutdown after ovf_properties injection.\n")
		}
	}

	//
	//  Grow boot disk to boot_disk_size
	//
	boot_disk_vmdkPATH, _ = getBootDiskPath(c, vmid)

	err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
	if err != nil {
		return vmid, fmt.Errorf("Failed to grow boot disk: %s\n", err)
	}

	//
	//  make updates to vmx file
	//
	err = updateVmx_contents(c, vmid, true, memsize, numvcpus, virthwver, guestos, virtual_networks, virtual_disks, notes, guestinfo)
	if err != nil {
		return vmid, fmt.Errorf("Failed to update vmx contents: %s\n", err)
	}

	return vmid, nil
}
