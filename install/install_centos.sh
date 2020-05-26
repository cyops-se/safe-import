#! /usr/bin/bash
# Copyright (c) 2020 Trailing bits AB

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
# EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
# MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
# IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
# DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
# OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
# OR OTHER DEALINGS IN THE SOFTWARE.

if [ "$EUID" -ne 0 ]
  then echo "This installation script must be started as root!"
  exit
fi

chmod +x ./bash-logger.sh
chmod +x ./install_si_centos.sh
source ./bash-logger.sh

function setup_users() {
    adduser si
}

function setup_filesystem() {
    INFO "Setting up file system structures and permissions"
    mkdir -p /safe-import/data/outer
    mkdir -p /safe-import/data/inner
    mkdir -p /safe-import/data/quarantine
    chown -R si.si /safe-import
}

function install_basic() {
    INFO "Installing basic packages"
    yum update
    yum -y install epel-release
    yum install -y git bubblewrap nodejs python3
    yum groupinstall -y "Development Tools"
}

function install_clamav() {
    INFO "Installing ClamAV"
    yum -y install clamav-data clamav-update clamav

    sestatus
    setsebool -P antivirus_can_scan_system 1
    setsebool -P clamd_use_jit 1
    sed -i -e "s/^Example/#Example/" /etc/clamd.d/scan.conf
    sed -i -e "s/#LocalSocket /LocalSocket /" /etc/clamd.d/scan.conf
    sed -i -e "s/^Example/#Example/" /etc/freshclam.conf
    freshclam

    cat <<EOD >/usr/lib/systemd/system/freshclam.service
[Unit]
Description = freshclam scanner
After = network.target

[Service]
Type = forking
ExecStart = /usr/bin/freshclam -d -c 1
Restart = on-failure
PrivateTmp =true

[Install]
WantedBy=multi-user.target
EOD
}

function install_gloang() {
    INFO "Installing Go"
    $pwd=`pwd`
    mkdir -p ~/download
    cd ~/download
    wget "https://dl.google.com/go/go1.14.3.linux-amd64.tar.gz"
    tar -xzf go1.14.3.linux-amd64.tar.gz
    mv go /usr/local
    export GOROOT=/usr/local/go
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
    echo "export GOROOT=/usr/local/go" >> /etc/profile
    echo "export PATH=$GOPATH/bin:$GOROOT/bin:$PATH" >> /etc/profile
    cd $pwd
}

function setup_firewall() {
    INFO "Setting up firewall zones and basic rules"
    firewall-cmd --new-zone=outer --permanent
    firewall-cmd --new-zone=inner --permanent
    #firewall-cmd --zone=outer --change-interface=eth1
    firewall-cmd --zone=inner --change-interface=eth0
    firewall-cmd --runtime-to-permanent
}

function install_safe_import() {
    INFO "Invoking user specific tasks"
    cp -f *.sh ~si/
    chown si.si ~si/*.sh
    sudo -iu si sh ./install_si_centos.sh
}

function setup_nginx() {
    INFO "Setting up NGINX"
    setsebool -P httpd_can_network_connect on
    chcon -Rt httpd_sys_content_t /home/si/run/si-webui
    chmod a+rx /home
    chmod a+rx /home/si
    chmod a+rx -R /home/si/run
    firewall-cmd --zone=inner --add-service=http --permanent
    firewall-cmd --reload
}

function final_setup() {
    INFO "Finalizing setup"
    # Allow safe-import uSvcs to bind DNS ports
    setcap 'cap_net_bind_service=+ep' /home/si/run/bin/server
}

setup_users
setup_filesystem
install_basic
install_clamav
install_golang
setup_firewall
install_safe_import
setup_nginx
final_setup