package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cyops-se/safe-import/usvc"
)

func main() {
	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	broker.SetTimeout(30)

	svc := &ChatService{}
	svc.Initialize(broker)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	s := <-c
	svc.LogGeneric("info", "Recevied signal: %v", s)
}
