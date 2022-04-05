#! /bin/sh
GIT_COMMIT=`git rev-list --abbrev-commit -1 HEAD`
GIT_VERSION=`git describe --tags --dirty --always`
CWD=`pwd`
cd $CWD/si-outer
go build -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"
cd $CWD/si-inner
go build -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"
cd $CWD/si-gatekeeper
go build -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"
cd $CWD/si-engine
go build -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION"