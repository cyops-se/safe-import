# Installation

- Full install through one single script if possible (specific distro as target)
- No installation options, script has full control
- Pre-requisites must be clear
- Check pre-requisites on startup

# Updates in focus
- Windows 10
- CentOS 7.6.1810
- pfsense 2.4.4-release-p2
- Suricata 4.1.2_3
- Snort 3.2.9.8_5
- LibreNMS 
- Moloch
- Synology
- Veeam?
- Custom upstream url
- Capture DNS for learning/blocking/detection


# Required packages
## Production
    // coredns (golang) - låtsas vara github och andra
    // dnsmasq - /etc/dnsmasq.conf address=/#/[ip address]
    git
    nginx
    nats.io
    postgres?
    clamav
    cuckoo?
    firejail
    cfssl

## Development
    yum groupinstall "Development Tools"
    golang - https://dl.google.com/go/go1.14.3.linux-amd64.tar.gz
    // coredns - git clone https://github.com/coredns/coredns

# Sandboxing
    https://lwn.net/Articles/686113/

# Anti-Virus
    ClamAV
    https://www.transip.eu/knowledgebase/entry/700-installing-clamav-in-centos7/
    yum install -y clam clam-update

# NATS
    https://github.com/nats-io/nats-server/releases/download/v2.1.7/nats-server-v2.1.7-linux-amd64.zip

# Configs
    https://www.tecmint.com/security-and-hardening-centos-7-guide/
    /dev/sda5 	/data        ext4    defaults,nosuid,nodev,noexec 1 2
    /dev/sda6  	/tmp         ext4    defaults,nosuid,nodev,noexec 0 0

https://www.howtoforge.com/samba-server-installation-and-configuration-on-centos-7


# safe-import software
- Web UI, Angular
- Engine, NATS usvcs
- bubblewrap

# Outer
- list of softwares to monitor and update
 - name, version, url, periodicity
 - check with http, https, smb, git

# Web UI
## Capabilities
- manage software list
- manage outer and inner storage areas
- 