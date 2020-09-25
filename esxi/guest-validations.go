package esxi

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func validateVirtualDiskSlot(slot string) string {
	log.Printf("[validateVirtualDiskSlot]\n")

	var result string

	// Split on comma.
	fields := strings.Split(slot+":UnSet", ":")

	// if using simple format
	if fields[1] == "UnSet" {
		fields[1] = fields[0]
		fields[0] = "0"
	}

	field0i, _ := strconv.Atoi(fields[0])
	field1i, _ := strconv.Atoi(fields[1])
	result = "ok"

	if field0i < 0 || field0i > 3 {
		result = "scsi controller id out of range"
	}
	if field1i < 0 || field1i > 15 {
		result = "scsi id out of range"
	}
	if field0i == 0 && field1i == 0 {
		result = "scsi id used by boot disk"
	}
	if field1i == 7 {
		result = "scsi id 7 not allowed"
	}

	return result
}

func validateNICType(nictype string) bool {
	log.Printf("[validateNICType]\n")

	if nictype == "" {
		return true
	}

	allNICtypes := `
    vlance
    flexible
    e1000
    e1000e
    vmxnet
    vmxnet2
    vmxnet3
	  `
	nictype = fmt.Sprintf(" %s\n", nictype)
	return strings.Contains(allNICtypes, nictype)
}

//func validateVirtHWver(guestos string) string {
//	log.Printf("[validateVirtHWver]\n")

//  return ""
//}

func validateGuestOsType(guestos string) bool {
	log.Printf("[validateGuestOsType]\n")

	if guestos == "" {
		return true
	}

	//  All valid Guest OS's
	allGuestOSs := [...]string{"asianux",
		"centos",
		"coreos",
		"darwin",
		"debian",
		"dos",
		"ecomstation",
		"fedora",
		"freebsd",
		"genericlinux",
		"mandrake",
		"mandriva",
		"netware",
		"nld9",
		"oes",
		"openserver",
		"opensuse",
		"oraclelinux",
		"os2",
		"other24xlinux",
		"other26xlinux",
		"other3xlinux",
		"other",
		"otherguest",
		"otherlinux",
		"redhat",
		"rhel",
		"sjds",
		"sles",
		"solaris",
		"suse",
		"turbolinux",
		"ubuntu",
		"unixware",
		"vmkernel",
		"vmwarephoton",
		"win31",
		"win95",
		"win98",
		"windows",
		"windowshyperv",
		"winlonghorn",
		"winme",
		"winnetbusiness",
		"winnetdatacenter",
		"winnetenterprise",
		"winnetstandard",
		"winnetweb",
		"winnt",
		"winvista",
		"winxphome",
		"winxppro",
	}

	guestos = fmt.Sprintf("%s\n", strings.ToLower(guestos))
	for i := 0; i < len(allGuestOSs); i++ {
		if strings.Contains(guestos, allGuestOSs[i]) {
			return true
		}
	}
	return false
}

func validateSCSIType(scsitype string) bool {
	log.Printf("[validateSCSIType]\n")

	if scsitype == "" {
		return true
	}

	allSCSItypes := `
    lsilogic
    pvscsi
    lsisas1068
	  `
	scsitype = fmt.Sprintf(" %s\n", scsitype)
	return strings.Contains(allSCSItypes, scsitype)
}
