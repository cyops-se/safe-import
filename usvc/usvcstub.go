package usvc

import (
	// "bufio"
	// "errors"

	"fmt"

	// "os"

	// "qtagg/ecosystem/contracts/system"
	// "time"

	"github.com/cyops-se/safe-import/usvc/types"
	"github.com/nats-io/nats.go"
)

type IUsvcStub interface {
	Name() string
	Component() string
	Version() int
	Fullname() string
}

type UsvcStub struct {
	name      string
	component string
	version   int
	fullname  string
	broker    *UsvcBroker
}

func CreateStub(broker *UsvcBroker, name string, component string, version int) *UsvcStub {
	stub := &UsvcStub{}
	stub.name = name
	stub.component = component
	stub.version = version
	stub.fullname = fmt.Sprintf("%d.%s.%s", stub.version, stub.component, stub.name)
	stub.broker = broker
	stub.broker.RegisterDependencies([]string{stub.fullname})
	return stub
}

func (stub *UsvcStub) Request(name string) (*types.Response, error) {
	if !stub.broker.IsServiceAvailable(stub.fullname) {
		return nil, fmt.Errorf("service '%s' not available at the moment", stub.fullname)
	}

	subject := fmt.Sprintf("methods.%s.%s", stub.fullname, name)
	return stub.broker.Request(subject)
}

func (stub *UsvcStub) RequestMessage(name string, request interface{}) (*types.Response, error) {
	if !stub.broker.IsServiceAvailable(stub.fullname) {
		return nil, fmt.Errorf("service '%s' not available at the moment", stub.fullname)
	}

	subject := fmt.Sprintf("methods.%s.%s", stub.fullname, name)
	return stub.broker.RequestMessage(subject, request)
}

func (stub *UsvcStub) PublishMessage(subject string, message interface{}) error {
	return stub.broker.PublishMessage(subject, &message)
}

func (stub *UsvcStub) PublishDataMessage(name string, message *interface{}) error {
	subject := fmt.Sprintf("data.%s.%s", stub.fullname, name)
	return stub.broker.PublishMessage(subject, message)
}

func (stub *UsvcStub) PublishEventMessage(name string, message *interface{}) error {
	subject := fmt.Sprintf("events.%s.%s", stub.fullname, name)
	return stub.broker.PublishMessage(subject, message)
}

func (stub *UsvcStub) PublishEventString(name string, json string) error {
	subject := fmt.Sprintf("events.%s.%s", stub.fullname, name)
	return stub.broker.PublishString(subject, json)
}

func (stub *UsvcStub) SubscribeEvent(name string, callback func(msg *nats.Msg)) error {
	subject := fmt.Sprintf("events.%s.%s", stub.fullname, name)
	return stub.broker.Subscribe(subject, callback)
}

func (stub *UsvcStub) SubscribeData(name string, callback func(msg *nats.Msg)) error {
	subject := fmt.Sprintf("data.%s.%s", stub.fullname, name)
	return stub.broker.Subscribe(subject, callback)
}
