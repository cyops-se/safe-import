package types

import "time"

const (
	ServiceState_INITIALIZING = 1
	ServiceState_ABORTING     = 2
	ServiceState_STARTING     = 3
	ServiceState_STOPPING     = 5
	ServiceState_PAUSING      = 6
	ServiceState_RUNNING      = 7
	ServiceState_STOPPED      = 8
	ServiceState_ABORTED      = 9
	ServiceState_PAUSED       = 10
	ServiceState_IDLE         = 11
)

type UsvcState struct {
	Name       string    `json:"name"`
	GitVersion string    `json:"gitversion"`
	Status     int       `json:"status"`
	LastSeen   time.Time `json:"lastseen"`
}

type MethodInfo struct {
	Name string `json:"name"`
}

type MethodInfos []*MethodInfo

type ServiceMetaResponse struct {
	ServiceInfos []*ServiceInfo `json:"infos"`
	State        ServiceState   `json:"state"`
}

type ServiceInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	MethodInfos MethodInfos  `json:"methods"`
	State       ServiceState `json:"state"`
}

type ServiceState uint

type Settings map[string]interface{}
