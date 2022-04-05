package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyops-se/safe-import/si-gatekeeper/common"
	"github.com/cyops-se/safe-import/si-gatekeeper/services"
	"github.com/cyops-se/safe-import/usvc"
)

var GitVersion string
var GitCommit string

// NOTE! This application violates service contracts internally as models defined by one usvc is
// directly accessed by others without invoking the required usvc interface. This construct is
// not recommended and should be considered for future refactoring
func main() {
	version := flag.Bool("v", false, "Prints the commit hash and exits")
	flag.Parse()

	usvc.GitVersion = GitVersion
	usvc.GitCommit = GitCommit

	if *version {
		fmt.Printf("si-gatekeeper version %s, commit %s\n", usvc.GitVersion, usvc.GitCommit)
		return
	}

	common.ConnectDatabase()

	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	broker.SetTimeout(30)

	proxysvc := &services.ProxyService{}
	proxysvc.Initialize(broker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	broker.LogGeneric("si-gatekeeper", "info", "Got request to exit")
}
