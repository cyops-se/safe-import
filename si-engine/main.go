package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyops-se/safe-import/si-engine/web/admin"
	"github.com/cyops-se/safe-import/si-engine/web/admin/routes"
	"github.com/cyops-se/safe-import/usvc"
)

var GitVersion string
var GitCommit string

func main() {
	version := flag.Bool("v", false, "Prints the commit hash and exits")
	flag.Parse()

	routes.SysInfo.GitVersion = GitVersion
	routes.SysInfo.GitCommit = GitCommit

	if *version {
		fmt.Printf("si-engine version %s, commit %s\n", routes.SysInfo.GitVersion, routes.SysInfo.GitCommit)
		return
	}

	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	broker.SetTimeout(30)

	go admin.Run(broker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// fmt.Println("Exit signal received... exiting")
}
