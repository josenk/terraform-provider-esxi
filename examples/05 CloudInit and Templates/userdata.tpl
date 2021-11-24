#cloud-config

packages:
 - ntp
 - ntpdate


# Add yum repository configuration to the system
yum_repos:
    # The name of the repository
    epel-testing:
        # Any repository configuration options
        # See: man yum.conf
        #
        # This one is required!
        baseurl: http://download.fedoraproject.org/pub/epel/testing/5/$basearch
        enabled: false
        failovermethod: priority
        gpgcheck: true
        gpgkey: file:///etc/pki/rpm-gpg/RPM-GPG-KEY-EPEL
        name: Extra Packages for Enterprise Linux 5 - Testing

runcmd:
    - date >/root/cloudinit.log
    - hostnamectl set-hostname ${HOSTNAME}
    - echo ${HELLO} >>/root/cloudinit.log
    - echo "Done cloud-init" >>/root/cloudinit.log
