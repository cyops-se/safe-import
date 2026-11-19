package main

import (
	"flag"
	"fmt"
	"log"
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

	cfg, err := admin.LoadConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	broker := &usvc.UsvcBroker{}
	if err := broker.Initialize(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	broker.SetTimeout(30)

	go func() {
		if err := admin.Run(broker, cfg); err != nil {
			log.Fatalf("Admin web server failed: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	broker.Shutdown()
}
