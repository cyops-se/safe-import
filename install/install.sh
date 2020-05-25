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

# if [ "$EUID" -ne 0 ]
#   then echo "This installation script must be started as root!"
#   exit
# fi

#sudo yum -y update
#sudo yum -y install epel-release

source bash-logger.sh

function check_install {
    result = `sudo yum list --installed | grep $1`
    return result
}

function setup_filesystem() {
    sudo mkdir -p /data/outer
    sudo mkdir -p /data/inner
    sudo mkdir -p /data/quarantine
    sudo mkdir -p /data/install
}

function install_basic() {
    sudo yum install -y git bubblewrap
}

function install_clamav() {    
    #clamav
    installed = check_install clamav
    
    sudo yum -y install clamav-data clamav-update clamav

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

function setup_users() {
    sudo adduser si
    sudo chown -R si.si /data
}

function install_nats() {
    wget https://github.com/nats-io/nats-server/releases/download/v2.1.7/nats-server-v2.1.7-linux-amd64.zip
    unzip nats-server-v2.1.7-linux-amd64.zip
    mv nats-server-v2.1.7-linux-amd64 nats-server
    rm nats-server-v2.1.7-linux-amd64.zip
    echo "NATS installed"
}

function download_safe_import() {
    wget https://github.com/nats-io/nats-server/releases/download/v2.1.7/nats-server-v2.1.7-linux-amd64.zip
}


setup_filesystem
setup_users
install_basic
install_clamav
download_safe_import
check_safe_import
build_safe_import