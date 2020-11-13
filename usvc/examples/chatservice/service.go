package main

import (
	"fmt"
	"log"

	"github.com/cyops-se/safe-import/usvc"
)

type ChatService struct {
	usvc.Usvc
}

func (svc *ChatService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "examples", "chat", "Basic chat service example")
	svc.RegisterMethod("history", svc.getHistory)
	svc.RegisterMethod("latest", svc.getLatest)
	svc.RegisterMethod("msg", svc.message)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()
}

func (svc *ChatService) execute() {
}

func (svc *ChatService) getHistory(payload string) (interface{}, error) {
	return nil, fmt.Errorf("Method not yet implemented: getHistory()")
}

func (svc *ChatService) getLatest(payload string) (interface{}, error) {
	return nil, fmt.Errorf("Method not yet implemented: getLatest()")
}

func (svc *ChatService) message(payload string) (interface{}, error) {
	log.Println("RECEIVED MESSAGE:", payload)
	svc.PublishDataString("msg", payload)
	return nil, nil
}
