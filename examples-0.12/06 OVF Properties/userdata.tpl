#cloud-config

packages:
 - ntp
 - ntpdate
 - curl

# Override ntp with chrony configuration on Ubuntu
ntp:
  enabled: true
  ntp_client: chrony  # Uses cloud-init default chrony configuration

runcmd:
  - date >/root/cloudinit.log
  - echo ${HELLO} >>/root/cloudinit.log
  - echo "Done cloud-init" >>/root/cloudinit.log
