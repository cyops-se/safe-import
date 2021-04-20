package usvc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyops-se/safe-import/usvc/types"

	"github.com/nats-io/nats.go"
)

// See microservicebroker_struct_[win/linux].go for the actual UsvcBroker type declaration

func (broker *UsvcBroker) Initialize() error {
	host, _ := os.Hostname()
	broker.hostname = fmt.Sprintf("%s.%d", host, os.Getpid())
	broker.timeout = 2

	broker.CheckConnection()
	broker.services = make(map[string]Usvc)
	broker.servicestates = make(map[string]*types.UsvcState)
	if broker.err != nil {
		broker.LogError("BROKER", "Broker failed to connect with NATS:", broker.err)
	}

	_, broker.err = broker.connection.Subscribe("system.heartbeat", broker.microServiceStateHandler)

	return broker.err
}

func (broker *UsvcBroker) RegisterDependencies(deps []string) {
	if broker.dependencies == nil {
		broker.dependencies = deps
	} else {
		broker.dependencies = append(broker.dependencies, deps...)
	}
}

func (broker *UsvcBroker) Shutdown() error {
	if broker.connection != nil {
		broker.connection.Close()
	}
	return nil
}

func (broker *UsvcBroker) Error() error {
	return broker.err
}

func (broker *UsvcBroker) SetTimeout(tmo uint) {
	broker.timeout = tmo
}

func (broker *UsvcBroker) Trace(fullname string, traceid string, msg string) {
	fullname = strings.ToLower(fullname)
	traceid = strings.ToLower(traceid)
	subject := fmt.Sprintf("trace.%s.%s", fullname, traceid)

	t := time.Now().UTC()
	json := fmt.Sprintf("{\"timestamp\":\"%s\", \"hostname\": \"%s\", \"subject\": \"%s\", \"msg\": \"%s\"}", t.Format("15:04:05.999"), broker.hostname, subject, msg)
	broker.PublishJson(subject, json)
}

func (broker *UsvcBroker) Log(fullname string, category string, title string) {
	subject := fmt.Sprintf("log.%s.%s", fullname, category)
	category = strings.ToLower(category)
	t := time.Now().UTC().Format("2006-01-02 15:04:05.000")
	if category == "info" {
		fmt.Printf("%s\tInfo: %s\n", t, title)
	} else if category == "warning" {
		fmt.Printf("%s\tWarning: %s\n", t, title)
	} else {
		fmt.Printf("%s\tTitle: %s\n", t, title)
	}

	var metalog types.MetaLog
	metalog.Time = time.Now().UTC()
	metalog.Source = fullname
	metalog.Category = category
	metalog.Title = title

	// json := fmt.Sprintf("{\"time\": \"%s\", \"source\": \"%s\", \"category\": \"%s\", \"title\": \"%s\"}", t, fullname, category, title)
	j, _ := json.Marshal(&metalog)

	broker.PublishBytes(subject, j)
}

func (broker *UsvcBroker) LogDescription(fullname string, category string, title string, description string) {
	subject := fmt.Sprintf("log.%s.%s", fullname, category)
	category = strings.ToLower(category)
	t := time.Now().UTC().Format("2006-01-02 15:04:05.000")
	if category == "info" {
		fmt.Printf("%s\tInfo: %s\n", t, title)
	} else if category == "warning" {
		fmt.Printf("%s\tWarning: %s\n", t, title)
	} else {
		fmt.Printf("%s\tTitle: %s, Description: %s\n", t, title, description)
	}

	var metalog types.MetaLog
	metalog.Time = time.Now().UTC()
	metalog.Source = fullname
	metalog.Category = category
	metalog.Title = title
	metalog.Description = description

	// json := fmt.Sprintf("{\"time\": \"%s\", \"source\": \"%s\", \"category\": \"%s\", \"title\": \"%s\"}", t, fullname, category, title)
	j, _ := json.Marshal(&metalog)

	broker.PublishBytes(subject, j)
}

func (broker *UsvcBroker) LogInfo(fullname string, msg string) {
	broker.Log(fullname, "info", msg)
}

func (broker *UsvcBroker) LogWarning(fullname string, msg string) {
	broker.Log(fullname, "warning", msg)
}

func (broker *UsvcBroker) LogError(fullname string, msg string, err error) {
	event := fmt.Sprintf("%s: %v", msg, err)
	broker.Log(fullname, "error", event)
}

func (broker *UsvcBroker) LogGeneric(name string, category string, format string, args ...interface{}) {
	title := fmt.Sprintf(format, args...)
	broker.Log(name, category, title)
}

func (broker *UsvcBroker) LogInfection(name string, title string, format string, args ...interface{}) {
	description := fmt.Sprintf(format, args...)
	broker.LogDescription(name, "infection", title, description)
}

func GetClientName() string {
	ex, err := os.Executable()
	if err != nil {
		return fmt.Sprintf("golang pid %d", os.Getpid())
	}
	_, file := filepath.Split(ex)
	return file
}

func (broker *UsvcBroker) CheckConnection() error {
	broker.err = nil
	if broker.connection == nil {
		broker.connection, broker.err = nats.Connect(nats.DefaultURL,
			nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second), // Allow infinitely many reconnects
			nats.Name(GetClientName()))
		if broker.err == nil { // If a new connection has been established, resubscribe
			for _, svc := range broker.services {
				broker.SubscribeToUsvcCalls(svc)
			}
		}
	}

	return broker.err
}

// Common code for RegisterUsvc and CheckConnection. Does not verify connection - be wary
func (broker *UsvcBroker) SubscribeToUsvcCalls(svc Usvc) error {
	subject := fmt.Sprintf("methods.%s.*", svc.Fullname())

	_, broker.err = broker.connection.Subscribe(subject, broker.microServiceV1Handler)
	if broker.err == nil {
		// broker.LogInfo("BROKER", fmt.Sprintf("UsvcBroker registering new service '%s'", dispatchname))
	} else {
		broker.LogWarning("BROKER", fmt.Sprintf("UsvcBroker failed to subscribe to '%s'", subject))
	}
	return broker.err
}

func (broker *UsvcBroker) RegisterUsvc(svc *Usvc) error {
	broker.err = broker.CheckConnection()

	dispatchname := "methods." + svc.Fullname()
	broker.services[dispatchname] = *svc
	if broker.err == nil {
		broker.SubscribeToUsvcCalls(*svc)
	}

	return broker.err
}

func (broker *UsvcBroker) DispatchBytes(subject string, data []byte) (interface{}, error, bool) {
	var request types.Request
	if broker.err = json.Unmarshal(data, &request); broker.err != nil {
		broker.LogError("BROKER", "UsvcBroker failed to unmarshal request message; ", broker.err)
		return nil, broker.err, false
	}

	result, err := broker.DispatchMessage(subject, &request)
	if err != nil {
		broker.err = err
	}

	return result, broker.err, false
}

func (broker *UsvcBroker) DispatchMessage(subject string, request *types.Request) (interface{}, error) {
	divindex := strings.LastIndex(subject, ".")
	method := subject[divindex+1:]
	name := subject[0:divindex]

	svc := broker.services[name]

	result, err := svc.DispatchLocalInvocation(method, request.Payload)
	if err != nil {
		broker.LogError("BROKER", "UsvcBroker received error from local service dispatch:", err)
		broker.err = err
	}

	return result, broker.err
}

func (broker *UsvcBroker) PublishBytes(subject string, message []byte) error {
	if broker.CheckConnection() == nil {
		broker.err = broker.connection.Publish(subject, message)
	}

	return broker.err
}

func (broker *UsvcBroker) PublishJson(subject string, message string) error {
	if broker.CheckConnection() == nil {
		broker.err = broker.PublishBytes(subject, []byte(message))
	}

	return broker.err
}

func (broker *UsvcBroker) PublishString(subject string, message string) error {
	if broker.CheckConnection() == nil {
		broker.err = broker.PublishBytes(subject, []byte(message))
	}

	return broker.err
}

func (broker *UsvcBroker) PublishMessage(subject string, message *interface{}) error {
	if broker.CheckConnection() == nil {
		data, _ := json.Marshal(message)
		broker.err = broker.PublishBytes(subject, data)
	}

	return broker.err
}

func (broker *UsvcBroker) RequestMessage(subject string, message interface{}) (*types.Response, error) {
	if broker.CheckConnection() == nil {
		payload, err := json.Marshal(message)
		request := &types.Request{Payload: string(payload)}

		data, _ := json.Marshal(request)
		rmsg, err := broker.connection.Request(subject, data, time.Duration(broker.timeout)*time.Second)
		if err == nil {
			response := &types.Response{}
			json.Unmarshal(rmsg.Data, response)
			if response.Header.Status.ResultCode > 0 {
				err = fmt.Errorf("Remote service reported an error: %d %s", response.Header.Status.ResultCode, response.Header.Status.ResultMessage)
			}
			return response, err
		}

		broker.err = err
	}

	return nil, broker.err
}

func (broker *UsvcBroker) Request(subject string) (*types.Response, error) {
	if broker.CheckConnection() == nil {
		request := &types.Request{}

		data, _ := json.Marshal(request)
		rmsg, err := broker.connection.Request(subject, data, time.Duration(broker.timeout)*time.Second)
		if err == nil {
			response := &types.Response{}
			json.Unmarshal(rmsg.Data, response)
			if response.Header.Status.ResultCode > 0 {
				err = fmt.Errorf("Remote service reported an error: %d %s", response.Header.Status.ResultCode, response.Header.Status.ResultMessage)
			}
			return response, err
		}

		broker.err = err
	}

	return nil, broker.err
}

func (broker *UsvcBroker) Subscribe(subject string, callback func(m *nats.Msg)) error {
	if broker.CheckConnection() == nil {
		_, broker.err = broker.connection.Subscribe(subject, callback)
		return broker.err
	}

	return broker.err
}

func (broker *UsvcBroker) IsServiceAvailable(fullname string) bool {
	if state, ok := broker.servicestates[fullname]; ok {
		return state.Status == types.ServiceState_RUNNING
	}

	return false
}

func (broker *UsvcBroker) microServiceV1Handler(msg *nats.Msg) {
	dispatchname := msg.Subject
	// // fmt.Println("microServiceV1Handler: received subject:", dispatchname);

	request := msg.Data
	response := &types.Response{}
	response.Header.FromHost = broker.hostname

	result, err, wrongHost := broker.DispatchBytes(dispatchname, request)
	if err == nil {
		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				broker.LogError("BROKER", "Failed to marshal result from "+dispatchname+" (request dispatch)", err)
				response.Header.Status.ResultCode = 99
				response.Header.Status.ResultMessage = err.Error()
			} else {
				response.Payload = string(data)
			}
		}
	} else {
		broker.LogError("BROKER", "Either a broker-internal error, a microservice failure or a bad client request happened in "+dispatchname+" (request dispatch)", err)
		response.Header.Status.ResultCode = 99
		response.Header.Status.ResultMessage = err.Error()
	}
	if wrongHost {
		return
	}
	data, _ := json.Marshal(response)
	broker.err = broker.connection.Publish(msg.Reply, data)
}

func (broker *UsvcBroker) microServiceStateHandler(msg *nats.Msg) {
	heartbeat := &types.Heartbeat{}
	if err := json.Unmarshal(msg.Data, heartbeat); err == nil {
		if entry, ok := broker.servicestates[heartbeat.Name]; ok {
			entry.LastSeen = time.Now().UTC()
			entry.Status = types.ServiceState_RUNNING
		} else {
			broker.servicestates[heartbeat.Name] = &types.UsvcState{Name: heartbeat.Name, GitVersion: heartbeat.GitVersion}
			broker.servicestates[heartbeat.Name].LastSeen = time.Now().UTC()
			broker.servicestates[heartbeat.Name].Status = types.ServiceState_RUNNING
		}
	}
}
