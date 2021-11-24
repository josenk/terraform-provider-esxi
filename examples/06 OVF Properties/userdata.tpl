#cloud-config


#  Install stuff
packages:
 - ntp
 - ntpdate
 - curl


# Override ntp with chrony configuration on Ubuntu
ntp:
  enabled: true
  ntp_client: chrony  # Uses cloud-init default chrony configuration


# Configure ubuntu user security
users:
  - name: ubuntu
    sudo: [ "ALL=(ALL) NOPASSWD:ALL" ]
    ssh-authorized-keys:
      - ssh-rsa AAAAB0000FAKE0SSH0AUTHORIZED0KEY00000iv0licRxqvoAyoRfLySGBs7f67LjyW2wZRaF/eyWTdUy2k6YKFEjlDL8iCbk3MsNMrd7U9ZitJnuME2Rw4mcR45XrYxSmZqMVRioBelvYmxzr6JhxY/zrr0tW4IFuc6VcYmyAwf0vHRzpYRzphP1JTKv63XhZtlpaFOvBv7LKRUooeHlhu4glw3wc7lXfXQbgHUeEiW8RmdkwR91YlkhEXWKihx2Q58uww+N846IilLP3i293nxPxmlAoSv17WsV+ZVkcBeJB2OLXdQxJDMowI8EwzsgDsMhHL/FHC58eyiCgD4M4TyQCINOn2U+SJZPBJz2YdW03HguG+EGqd8HJFauwX9nJSwnKvcgjt2L5oTaF2++oEGTc6tMHizMPGBvsYGCa6LwgkDuxHmSa69SX0Okth1QZQOnEH731jgED/pBJuVcrRcyPUQRNSV1GQRh9ZOnDCE2YKAD89jB66Wk+INMvFftQ3PiIxedK//ahFKW1XN8YZrjY/kouEFRMIHfKUYu/SILBJSNNKMBbBpuvSo5N8t15Nfrfh3n+mY4TroR1ASOWMILQ/M5BxFn70uUOEokmvUaOZZlJZ3YPo+0lZadCLiYEGqvDJtE2UyKfFQTIw/udIvl+P3fnFRAEfxdd66dxf2pUt18X2c6qhQ0WD1SMveeEA12bzh3w== ubuntu@dev


#  Change some default passwords
chpasswd:
  list: |
    root:ubuntu1
    ubuntu:ubuntu2
  expire: False


#  Write to a log file (useing variables set in terraform) and show the ip on the console.
runcmd:
  - date >/root/cloudinit.log
  - hostnamectl set-hostname ${HOSTNAME}
  - echo ${HELLO} >>/root/cloudinit.log
  - echo "Done cloud-init" >>/root/cloudinit.log
  - ip a >/dev/tty1
  
