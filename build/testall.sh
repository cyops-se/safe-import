#! /bin/sh
GIT=COMMIT=`git rev-list --abbrev-commit -1 HEAD`
GIT_VERSION=`git describe --tags --dirty --always`
nats-server &
sleep 1
CWD=`pwd`
cd $CWD/si-outer
./si-outer &
sleep 1
cd $CWD/si-inner
sudo ./si-inner &
sleep 1
cd $CWD/si-gatekeeper
./si-gatekeeper &
sleep 1
cd $CWD/si-engine
./si-engine