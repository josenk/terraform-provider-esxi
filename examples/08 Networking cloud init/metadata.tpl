network:
    version: 2
    ethernets:
        ens192:
            dhcp4: false
            addresses:
                - ${ipAddress}
            gateway4: ${gateway}
            nameservers:
                addresses:
                    - ${nameserver}

# example
# network:
#     version: 2
#     ethernets:
#         ens192:
#             dhcp4: false
#             addresses:
#                 - 10.10.10.1/24
#             gateway4: 10.10.10.254
#             nameservers:
#                 addresses:
#                     - 8.8.8.8