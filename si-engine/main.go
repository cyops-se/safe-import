package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyops-se/safe-import/si-engine/web/admin"
	"github.com/cyops-se/safe-import/usvc"
)

func main() {
	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	broker.SetTimeout(30)

	go admin.Run(broker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Exit signal received... exiting")
}
