#! /bin/sh
GIT=COMMIT=`git rev-list --abbrev-commit -1 HEAD`
GIT_VERSION=`git describe --tags --dirty --always`
go build -ldflags "-X main.GitCommit=%GIT_COMMIT% -X main.GitVersion=%GIT_VERSION%"