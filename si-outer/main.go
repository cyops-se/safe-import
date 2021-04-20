package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cyops-se/safe-import/si-outer/common"
	"github.com/cyops-se/safe-import/si-outer/services"
	"github.com/cyops-se/safe-import/si-outer/system"
	"github.com/cyops-se/safe-import/usvc"
)

func main() {
	common.ConnectDatabase()
	system.Init()

	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	broker.SetTimeout(30)

	jobsvc := &services.JobsService{}
	jobsvc.Initialize(broker)

	// reposvc := &services.RepoService{}
	// reposvc.Initialize(broker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	broker.LogGeneric("si-outer", "info", "Got request to exit")
}
