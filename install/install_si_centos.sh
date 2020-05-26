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

if [ "$EUID" -eq 0 ]
  then echo "This installation script must NOT be started as root!"
  exit
fi

source ./bash-logger.sh

function download_safe_import() {
    mkdir -p run/bin
    mkdir download

    export GIT_TERMINAL_PROMPT=1

    INFO Downloading and installing NATS.o server
    cd ~/download
    wget https://github.com/nats-io/nats-server/releases/download/v2.1.7/nats-server-v2.1.7-linux-amd64.zip
    unzip nats-server-v2.1.7-linux-amd64.zip
    mv nats-server-v2.1.7-linux-amd64/nats-server ~/run/bin

    INFO Downloading and installing safe-import backend
    cd ~/download
    #wget https://github.com/cyops-se/si-api/archive/master.zip -o si-api.zip
    #unzip si-api.zip
    #mv si-api-master ~/run/si-api
    git clone https://github.com/cyops-se/si-api
    mv si-api ~/run
    cd ~/run/si-api
    npm i
    
    INFO Downloading and installing safe-import web user interface
    cd ~/download
    #wget https://github.com/cyops-se/si-webui/archive/master.zip -o si-webui.zip
    #unzip si-webui.zip 'si-webui-master/dist/*'
    #mv si-webui-master/dist ~/run/si-webui
    git clone https://github.com/cyops-se/si-webui
    mv si-webui/dist ~/run/si-webui
    rm -rf si-webui-master

    INFO Downloading and installing safe-import micro services
    cd ~/run
    mkdir go
    export GOPATH=/home/si/run/go
    go get "github.com/cyops-se/si-usvc/server"
}

function check_safe_import() {
    INFO Checking safe-import sources (including dependencies) for malwarels 
    cd ~/run
    clamscan -ir
}

function build_safe_import() {
    INFO Building safe-import micro services
    cd ~/run/go/src/github.com/cyops-se/si-usvc/server
    go build
    mv server ~/run/bin
}

download_safe_import
check_safe_import
build_safe_import