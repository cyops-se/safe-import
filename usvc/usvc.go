package usvc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	runtimedebug "runtime/debug"
	"time"

	"github.com/cyops-se/safe-import/usvc/types"
)

var (
	GitVersion string
)

type Usvc struct {
	name           string
	component      string
	description    string
	version        int
	state          types.ServiceState
	duration       time.Duration
	ticker         *time.Ticker
	methods        map[string](func(payload string) (interface{}, error))
	broker         *UsvcBroker
	Settings       types.Settings
	Executor       func()
	ApplySettings  func()
	chChangeTicker chan bool
}

func (svc *Usvc) Name() string {
	return svc.name
}

func (svc *Usvc) Component() string {
	return svc.component
}

func (svc *Usvc) Version() int {
	return svc.version
}

func (svc *Usvc) Fullname() string {
	return fmt.Sprintf("%d.%s.%s", svc.version, svc.component, svc.name)
}

func (svc *Usvc) State() types.ServiceState {
	return svc.state
}

// duration in seconds
func (svc *Usvc) SetTaskIdleTime(durationSeconds int64) {
	svc.ticker.Stop()
	svc.duration = time.Duration(durationSeconds) * time.Second
	svc.ticker = time.NewTicker(svc.duration)
	select {
	case svc.chChangeTicker <- true: // Kick the goroutine, so it doesn't wait forever
		break
	default:
		break // Don't send if nobody is waiting
	}
}

func (svc *Usvc) InitializeService(broker *UsvcBroker, version int, component string, name string, description string) {
	svc.state = types.ServiceState_INITIALIZING // Probably unnecessary

	svc.methods = make(map[string](func(payload string) (interface{}, error)))
	svc.version = version
	svc.name = name
	svc.component = component
	svc.description = description
	svc.broker = broker
	svc.Executor = svc.defaultexecute
	svc.ApplySettings = svc.defaultapplysettings
	svc.chChangeTicker = make(chan bool)
	svc.Settings = make(map[string]interface{})

	svc.broker.RegisterUsvc(svc)
	svc.RegisterMethod("meta-info", svc.metaInfo)
	svc.RegisterMethod("meta-op", svc.metaOp)
	svc.RegisterMethod("meta-get", svc.metaGet)
	svc.RegisterMethod("meta-getall", svc.metaGetAll)
	svc.RegisterMethod("meta-set", svc.metaSet)
	svc.RegisterMethod("meta-setall", svc.metaSetAll)
	svc.RegisterMethod("meta-apply", svc.metaApply)

	svc.duration = 5 * time.Second
	svc.ticker = time.NewTicker(svc.duration)
	// svc.taskidletime = 5 // in seconds
	svc.state = types.ServiceState_STARTING
	go svc.jobengine()

	time.AfterFunc(time.Second, svc.heartbeat)
}

func (svc *Usvc) RegisterMethod(name string, callback func(payload string) (interface{}, error)) {
	// fmt.Printf("Method %s registered for service %s\n", name, svc.Fullname())
	svc.methods[name] = callback
}

func (svc *Usvc) DispatchLocalInvocation(method string, payload string) (interface{}, error) {
	defer svc.LogPanic(fmt.Sprintf("%s-%s", "DispatchLocalInvocation", method))
	callback := svc.methods[method]
	var result interface{}
	var err error

	if callback != nil {
		result, err = callback(payload)
	} else {
		err = fmt.Errorf("Usvc method handler not registered %s, %s", method, svc.Fullname())
	}

	return result, err
}

func (svc *Usvc) Trace(traceid string, msg string) {
	svc.broker.Trace(svc.Fullname(), traceid, msg)
}

func (svc *Usvc) LogInfo(msg string) {
	svc.broker.LogInfo(svc.Fullname(), msg)
}

func (svc *Usvc) LogWarning(msg string) {
	svc.broker.LogWarning(svc.Fullname(), msg)
}

func (svc *Usvc) LogError(msg string, err error) {
	svc.broker.LogError(svc.Fullname(), msg, err)
}

func (svc *Usvc) LogGeneric(category string, format string, args ...interface{}) {
	svc.broker.LogGeneric(svc.Fullname(), category, format, args...)
}

func (svc *Usvc) Publish(subject string, msg interface{}) error {
	return svc.broker.PublishMessage(subject, &msg)
}

func (svc *Usvc) PublishData(name string, msg interface{}) error {
	subject := fmt.Sprintf("data.%s.%s", svc.Fullname(), name)
	return svc.broker.PublishMessage(subject, &msg)
}

func (svc *Usvc) PublishDataString(name string, msg string) error {
	subject := fmt.Sprintf("data.%s.%s", svc.Fullname(), name)
	return svc.broker.PublishString(subject, msg)
}

func (svc *Usvc) PublishEventMessage(name string, msg interface{}) error {
	subject := fmt.Sprintf("events.%s.%s", svc.Fullname(), name)
	return svc.broker.PublishMessage(subject, &msg)
}

func (svc *Usvc) PublishEventString(name string, msg string) error {
	subject := fmt.Sprintf("events.%s.%s", svc.Fullname(), name)
	return svc.broker.PublishString(subject, msg)
}

func (svc *Usvc) Pause() {
	svc.state = types.ServiceState_PAUSING
}

func (svc *Usvc) Resume() {
	svc.state = types.ServiceState_STARTING
}

func (svc *Usvc) Stop() {
	svc.state = types.ServiceState_STOPPING
}

func (svc *Usvc) Start() {
	svc.state = types.ServiceState_STARTING
}

// This method doesn't always work - svc.state can be overwritten before the jobengine loop terminates
func (svc *Usvc) Abort() {
	svc.state = types.ServiceState_ABORTING
}

func (svc *Usvc) LoadSettings() error {
	filename := fmt.Sprintf("/cyops/usvc/settings/%s_%s.json", svc.component, svc.name)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		svc.LogError("Unable to load settings", err)
		return err
	}

	err = json.Unmarshal(data, &svc.Settings)
	if err != nil {
		svc.LogError("Unable to interpret settings", err)
	}

	return err
}

func (svc *Usvc) SaveSettings() {
	folderpath := "/cyops/usvc/settings"
	filename := fmt.Sprintf("%s/%s_%s.json", folderpath, svc.component, svc.name)

	// Create folder if it doesn't exist
	if _, err := os.Stat(folderpath); os.IsNotExist(err) {
		os.MkdirAll(folderpath, os.ModePerm)
	}

	content, err := json.Marshal(&svc.Settings)
	if err == nil {
		svc.LogInfo("Saving settings")
		file, _ := os.Create(filename)
		file.Write([]byte(content))
		file.Close()

		svc.ApplySettings()

		// Publish the new settings to anyone interested (as JSON)
		svc.PublishData("settings", &svc.Settings)
	} else {
		svc.LogError("Unable to save settings", err)
	}
}

// Service methods implementations
//

func (svc *Usvc) metaInfo(payload string) (interface{}, error) {
	var err error
	response := &types.ServiceMetaResponse{}
	svcinfo := &types.ServiceInfo{}

	svcinfo.Name = svc.name
	svcinfo.Description = svc.description
	svcinfo.State = types.ServiceState(svc.state)

	for k, _ := range svc.methods {
		methodinfo := &types.MethodInfo{Name: k} // TODO: Add description (probably needs a new struct for registering methods rather than a simple func)
		svcinfo.MethodInfos = append(svcinfo.MethodInfos, methodinfo)
	}

	response.ServiceInfos = append(response.ServiceInfos, svcinfo)

	return response, err
}

func (svc *Usvc) metaOp(payload string) (interface{}, error) {
	var err error
	var request types.MetaOpRequest
	if err = json.Unmarshal([]byte(payload), &request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal JSON request: metaOp(%s)", payload)
	}

	operation := request.Operation
	if operation == "stop" {
		svc.Stop()
	} else if operation == "start" {
		svc.Start()
	} else if operation == "pause" {
		svc.Pause()
	} else if operation == "resume" {
		svc.Resume()
	} else if operation == "exit" {
		os.Exit(0)
	}

	return nil, err
}

// Takes a String (Json.Text) as argument and returns SettingsItem
func (svc *Usvc) metaGet(payload string) (interface{}, error) {
	var err error
	var request types.MetaGetRequest
	if err = json.Unmarshal([]byte(payload), &request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal JSON request: metaGet(%s)", payload)
	}

	if val, ok := svc.Settings[request.Name]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("No such settings item: %s", request.Name)
}

func (svc *Usvc) metaGetAll(notused string) (interface{}, error) {
	return &svc.Settings, nil
}

func (svc *Usvc) metaSet(payload string) (interface{}, error) {
	var err error
	var request types.MetaSetRequest
	if err = json.Unmarshal([]byte(payload), &request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal JSON request: metaGet(%s)", payload)
	}

	svc.Settings[request.Name] = request.Value
	svc.SaveSettings()

	return nil, nil
}

func (svc *Usvc) metaSetAll(payload string) (interface{}, error) {
	if err := json.Unmarshal([]byte(payload), &svc.Settings); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal JSON request: metaSetAll(%s)", payload)
	}

	svc.SaveSettings()
	return nil, nil
}

func (svc *Usvc) metaApply(payload string) (interface{}, error) {
	svc.ApplySettings()
	return nil, nil
}

// Windows services can't print to console, so we have to explicitly handle any panics.
func (svc *Usvc) LogPanic(caller string) {
	if r := recover(); r != nil {
		msg := fmt.Sprintf("Panic in %s %s: %v\nGitVersion:%s\n%s", svc.Fullname(), caller, r, GitVersion, string(runtimedebug.Stack()))
		svc.LogError("Panic", fmt.Errorf("%s", msg))
		panic(msg) // Re-panic to cause microservice to crash so the supervisor restarts us
	}
}

func (svc *Usvc) jobengine() {
	defer svc.LogPanic("jobengine")
	svc.state = types.ServiceState_RUNNING
	for svc.state != types.ServiceState_ABORTING && svc.state != types.ServiceState_ABORTED {

		for {
			select {
			case <-svc.ticker.C:
				break
			case <-svc.chChangeTicker:
				continue
			}
			break
		}

		if svc.state == types.ServiceState_IDLE {
			svc.state = types.ServiceState_RUNNING
		}

		if svc.state == types.ServiceState_RUNNING {
			svc.execute()
			svc.state = types.ServiceState_IDLE
		} else if svc.state == types.ServiceState_PAUSING {
			svc.pause()
		} else if svc.state == types.ServiceState_STOPPING {
			svc.stop()
		} else if svc.state == types.ServiceState_STARTING {
			svc.state = types.ServiceState_RUNNING
		}
	}
	svc.state = types.ServiceState_ABORTED
	fmt.Printf("EXIT jobengine %s...\n", svc.Fullname())
}

func (svc *Usvc) pause() {
	svc.state = types.ServiceState_PAUSED
	for svc.state == types.ServiceState_PAUSED {
		time.Sleep(1 * time.Second)
	}
}

func (svc *Usvc) stop() {
	svc.state = types.ServiceState_STOPPED
	for svc.state == types.ServiceState_STOPPED {
		time.Sleep(1 * time.Second)
	}
}

func (svc *Usvc) defaultexecute() {
	// fmt.Printf("%s default executor activated\n", svc.Fullname())
}

func (svc *Usvc) defaultapplysettings() {
	// fmt.Printf("%s default executor activated\n", svc.Fullname())
}

func (svc *Usvc) execute() {
	if svc.Executor != nil {
		svc.Executor()
	}
}

func (svc *Usvc) heartbeat() {
	hb := &types.Heartbeat{Name: svc.Fullname(), Version: &types.Version{Major: svc.version, Minor: 0},
		CurrentUTC: time.Now().UTC(), Hostname: svc.broker.hostname, GitVersion: GitVersion}
	svc.Publish("system.heartbeat", hb)
	time.AfterFunc(time.Second, svc.heartbeat)
}
