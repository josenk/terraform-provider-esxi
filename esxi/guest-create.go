package esxi

import (
	"bytes"
	"fmt"
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
	boot_disk_type string, boot_disk_size string, virtual_networks [10][3]string, boot_firmware string,
	virtual_disks [60][2]string, guest_shutdown_timeout int, ovf_properties_timer int, notes string,
	guestinfo map[string]interface{}, ovf_properties map[string]string) (string, error) {

	esxiConnInfo := getConnectionInfo(c)

	var memsize, numvcpus, virthwver int
	var boot_disk_vmdkPATH, remote_cmd, vmid, stdout, vmx_contents string

	// Indicates that ovftool requires ovf properties to be passed through
	// (and thus requires slightly different way of managing disk resizing)
	usesOvfProperties := len(ovf_properties) > 0

	memsize, _ = strconv.Atoi(strmemsize)
	numvcpus, _ = strconv.Atoi(strnumvcpus)
	virthwver, _ = strconv.Atoi(strvirthwver)

	//
	//  Check if Disk Store already exists
	//
	err := diskStoreValidate(c, disk_store)
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
				fmt.Sprintf("scsi0:0.deviceType = \\\"scsi-hardDisk\\\"\n") +
				fmt.Sprintf("nvram = \\\"%s.nvram\\\"\n", guest_name)
		if boot_firmware == "efi" {
			vmx_contents = vmx_contents +
				fmt.Sprintf("firmware = \\\"efi\\\"\n")
		} else if boot_firmware == "bios" {
			vmx_contents = vmx_contents +
				fmt.Sprintf("firmware = \\\"bios\\\"\n")
		}
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
				fmt.Sprintf("ide1:0.deviceType = \\\"cdrom-raw\\\"\n")
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
		// Setup command line for ovftool invocation
		var args strings.Builder

		// Indicates that ovftool on ESXi host should be used
		useRemoteOvfTool := false

		// Add any OVF properties needed to command line; also note that in order for OVF properties to work,
		// VM must immediately be powered on (which we have to account for when changing disk size, for example)
		if usesOvfProperties {
			args.WriteString("--X:injectOvfEnv --powerOn ")
			for k, v := range ovf_properties {
				args.WriteString(fmt.Sprintf("--prop:%s=%s ", k, v))
			}
		}

		// Add any guest info flags into extra config settings
		if len(guestinfo) > 0 {
			args.WriteString("--allowExtraConfig ")
			for k, v := range guestinfo {
				args.WriteString(fmt.Sprintf("--extraConfig:guestinfo.%s=%s ", k, v))
			}
		}

		// Ensure source file exists. There are 4 variations:
		// * vi URL - source is a guest VM; nothing to verify
		// * HTTP URL - verify it's a 200 response (using HEAD)
		// * host_ovf - source is a path *on* the ESXi box (requires ovftool to be installed on ESXi box as well)
		// * Local file - source is a local file that will be uploaded
		if strings.HasPrefix(src_path, "vi://") {
			log.Printf("[guestCREATE] ovf_source is guest VM")

		} else if strings.HasPrefix(src_path, "http://") || strings.HasPrefix(src_path, "https://") {
			log.Printf("[guestCREATE] ovf_source is HTTP/S URL")
			resp, err := http.Head(src_path)
			defer resp.Body.Close()
			if (err != nil) || (resp.StatusCode != 200) {
				log.Printf("[guestCREATE] URL not accessible: %s\n", src_path)
				log.Printf("[guestCREATE] URL StatusCode: %d\n", resp.StatusCode)
				log.Printf("[guestCREATE] URL Error: %s\n", err.Error())
				return "", fmt.Errorf("ovf_source URL not accessible: %s\n%s", src_path, err.Error())
			}

		} else if strings.HasPrefix(src_path, "host_ovf://") {
			log.Printf("[guestCREATE] ovf_source is path on ESXi host")

			// Make sure remote OVF tool is defined; at this point, we've verified tool is present as well
			if c.esxiRemoteOvfToolPath == "" {
				return "", fmt.Errorf("host_ovf source configured, but no path found for ovftool on ESXi!")
			}
			useRemoteOvfTool = true
			src_path = strings.TrimPrefix(src_path, "host_ovf://")

		} else {
			log.Printf("[guestCREATE] ovf_source is local file\n")
			useRemoteOvfTool = false
			if _, err := os.Stat(src_path); os.IsNotExist(err) {
				log.Printf("[guestCREATE] File not found, Error: %s\n", err.Error())
				return "", fmt.Errorf("ovf_source file not found: %s\n", src_path)
			}
		}

		//  Set disk mode param
		if boot_disk_type == "zeroedthick" {
			boot_disk_type = "thick"
		}
		args.WriteString(fmt.Sprintf("--diskMode=%s ", boot_disk_type))

		// Construct destination path for ovftool
		username := url.QueryEscape(c.esxiUserName)
		password := url.QueryEscape(c.esxiPassword)
		dst_path := fmt.Sprintf("vi://%s:%s@%s:%s/%s", username, password, c.esxiHostName, c.esxiHostSSLport, resource_pool_name)

		// If the source is an OVA or OVF and a virtual network is defined, add parameter for ovftool
		if (strings.HasSuffix(src_path, ".ova") || strings.HasSuffix(src_path, ".ovf")) && virtual_networks[0][0] != "" {
			args.WriteString(fmt.Sprintf("--network='%s' ", virtual_networks[0][0]))
		}

		// Include guest name
		args.WriteString(fmt.Sprintf("--name=%s ", guest_name))

		// Include data store
		args.WriteString(fmt.Sprintf("--datastore=%s ", disk_store))

		// Add other parameters to ovftool
		args.WriteString("--acceptAllEulas --noSSLVerify ")

		// Finalize arguments to ovftool
		ovf_args := fmt.Sprintf("%s %s %s", args.String(), src_path, dst_path)

		// Log sanitized set of args to ovftool
		re := regexp.MustCompile("vi://.*?@")
		sanitized_ovf_args := re.ReplaceAllString(ovf_args, "vi://XXXX:YYYY@")
		log.Printf("[guestCREATE] Invoking ovftool with args: %s\n", sanitized_ovf_args)

		var ovftoolOutput string
		var ovftoolErr error

		if useRemoteOvfTool {
			// Invoke ovftool on ESXi host via SSH
			ovftoolOutput, ovftoolErr = runRemoteSshCommand(esxiConnInfo, fmt.Sprintf("%s %s", c.esxiRemoteOvfToolPath, ovf_args), "remote ovftool")
		} else {
			ovftoolOutput, ovftoolErr = runOvfTool(ovf_args)
		}

		log.Printf("[guestCREATE] ovftool output: %s\n", ovftoolOutput)
		if ovftoolErr != nil {
			log.Printf("[guestCREATE] failed to invoke remote ovftool: %s\n", ovftoolErr)
			return "", fmt.Errorf("failed to invoke remote ovftool: %s\n", ovftoolErr)
		}
	}

	// Get VMID (by name)
	vmid, err = guestGetVMID(c, guest_name)
	if err != nil {
		return "", fmt.Errorf("Failed to get vmid: %s\n", err)
	}

	// OVF properties require ovftool to power on the VM to inject the properties.
	// Unfortunately, there is no way to know when cloud-init is finished. So, we
	// wait for ovf_properties_timer seconds, then shutdown/power-off to continue and hope
	// system comes down cleanly.
	if usesOvfProperties == true {
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

	_, err = growVirtualDisk(c, boot_disk_vmdkPATH, boot_disk_size)
	if err != nil {
		return vmid, fmt.Errorf("Failed to grow boot disk: %s\n", err)
	}

	//
	//  make updates to vmx file
	//
	err = updateVmx_contents(c, vmid, true, memsize, numvcpus, virthwver, guestos, virtual_networks, boot_firmware, virtual_disks, notes, guestinfo)
	if err != nil {
		return vmid, fmt.Errorf("Failed to update vmx contents: %s\n", err)
	}

	return vmid, nil
}

func runOvfTool(ovf_args string) (string, error) {
	osShellCmd := "/bin/bash"
	osShellCmdOpt := "-c"

	ovf_cmd := "ovftool " + ovf_args

	// On Windows, we write a temporary batch file to invoke ovftool (why??)
	if runtime.GOOS == "windows" {
		osShellCmd = "cmd.exe"
		osShellCmdOpt = "/c"

		// Replace any single quotes with escaped double quotes and escape any percent signs
		ovf_cmd = strings.Replace(ovf_cmd, "'", "\"", -1)
		ovf_cmd = strings.Replace(ovf_cmd, "%", "%%", -1)

		// Create a temp file
		file, err := os.CreateTemp("", "ovf_cmd*.bat")
		if err != nil {
			return "", fmt.Errorf("failed to create temp ovf_cmd.bat file: %+v", err)
		}

		_, err = file.WriteString(ovf_cmd)
		if err != nil {
			file.Close()
			return "", fmt.Errorf("failed to write ovf_cmd.bat file: %+v", err)
		}

		err = file.Close()
		if err != nil {
			return "", fmt.Errorf("failed to close ovf_cmd.bat file: %+v", err)
		}

		ovf_cmd = file.Name()
	}

	var out bytes.Buffer
	cmd := exec.Command(osShellCmd, osShellCmdOpt, ovf_cmd)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to invoke local ovftool: %s\n", err)
	}

	return out.String(), nil
}
