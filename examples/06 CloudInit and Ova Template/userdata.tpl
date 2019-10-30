#cloud-config

users:
  - default
  - name: slavko
    passwd: $1$slavko$3qoKYqsEr9ZU9xhAiTTuB.
    ssh_pwauth: True
    chpasswd: { expire: False }
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC07jki/9uEq9Pn+hry3lmGjMeJORdETVrpFL53X0aZLJp8KvWcqFf+P/ZdScHnVqr0VEt+axtyivSf0TrfB/i97n9lW2ZA0RY9UnURJqPpknfvidgosDzqGTFlOfdvE/tw7QzK0G8AmyxaQ0pf4ueD0brk4k0IVWX7oMdT/rJMb05owDaj3E6LK5smeGe86i9R9oFmBDxKTD9CSTzG8T70MsaNR4/brBtDFpDyVFzScnNB9xN8xXnalFiJBtkqyGYXqshGPxWMzTAXv6Zmjxe/hDUfBT0wgp6yscj0BCippsFyZ+LhK/ChamtZAEceWveH06nTc/+Kh3c59RMo0MvP slavko@rocketracoon
packages:
 - ntp
 - ntpdate
 - curl

# Override ntp with chrony configuration on Ubuntu
ntp:
  enabled: true
  ntp_client: chrony  # Uses cloud-init default chrony configuration


# Add yum repository configuration to the system
#yum_repos:
#    # The name of the repository
#    epel-testing:
#        # Any repository configuration options
#        # See: man yum.conf
#        #
#        # This one is required!
#        baseurl: http://download.fedoraproject.org/pub/epel/testing/5/$basearch
#        enabled: false
#        failovermethod: priority
#        gpgcheck: true
#        gpgkey: file:///etc/pki/rpm-gpg/RPM-GPG-KEY-EPEL
#        name: Extra Packages for Enterprise Linux 5 - Testing

runcmd:
    - date > /tmp/cloudinit.log
    - whoami >> /tmp/cloudinit.log
    - sudo dhclient ens36
    - sudo hostnamectl set-hostname ${HOSTNAME}
    - sudo echo ${HELLO} >> /tmp/cloudinit.log
    - sudo echo "Done tf cloud-init" >>/tmp/cloudinit.log
