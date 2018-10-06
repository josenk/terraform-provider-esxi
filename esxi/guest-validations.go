package esxi

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	//"errors"
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
	allGuestOSs := `
	  asianux3-64
    asianux3
    asianux4-64
    asianux4
    asianux5-64
    asianux7-64
    centos6-64
    centos-64
    centos6
    centos7-64
    centos7
    centos
    coreos-64
    darwin10-64
    darwin10
    darwin11-64
    darwin11
    darwin12-64
    darwin13-64
    darwin14-64
    darwin15-64
    darwin16-64
    darwin-64
    darwin
    debian10-64
    debian10
    debian4-64
    debian4
    debian5-64
    debian5
    debian6-64
    debian6
    debian7-64
    debian7
    debian8-64
    debian8
    debian9-64
    debian9
    dos
    ecomstation2
    ecomstation
    fedora-64
    fedora
    freebsd-64
    freebsd
    genericlinux
    mandrake
    mandriva-64
    mandriva
    netware4
    netware5
    netware6
    nld9
    oes
    openserver5
    openserver6
    opensuse-64
    opensuse
    oraclelinux6-64
    oraclelinux-64
    oraclelinux6
    oraclelinux7-64
    oraclelinux7
    oraclelinux
    os2
    other24xlinux-64
    other24xlinux
    other26xlinux-64
    other26xlinux
    other3xlinux-64
    other3xlinux
    other
    otherguest-64
    otherlinux-64
    otherlinux
    redhat
    rhel2
    rhel3-64
    rhel3
    rhel4-64
    rhel4
    rhel5-64
    rhel5
    rhel6-64
    rhel6
    rhel7-64
    rhel7
    sjds
    sles10-64
    sles10
    sles11-64
    sles11
    sles12-64
    sles12
    sles-64
    sles
    solaris10-64
    solaris10
    solaris11-64
    solaris6
    solaris7
    solaris8
    solaris9
    suse-64
    suse
    turbolinux-64
    turbolinux
    ubuntu-64
    ubuntu
    unixware7
    vmkernel5
    vmkernel65
    vmkernel6
    vmkernel
    vmwarephoton-64
    win2000advserv
    win2000pro
    win2000serv
    win31
    win95
    win98
    windows7-64
    windows7
    windows7server-64
    windows8-64
    windows8
    windows8server-64
    windows9-64
    windows9
    windows9server-64
    windowshyperv
    winlonghorn-64
    winlonghorn
    winme
    winnetbusiness
    winnetdatacenter-64
    winnetdatacenter
    winnetenterprise-64
    winnetenterprise
    winnetstandard-64
    winnetstandard
    winnetweb
    winnt
    winvista-64
    winvista
    winxphome
    winxppro-64
    winxppro
		`

	guestos = fmt.Sprintf(" %s\n", guestos)
	return strings.Contains(allGuestOSs, guestos)
}
