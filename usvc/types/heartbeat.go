package types

import "time"

type Version struct {
	Minor int `json:"minor"`
	Major int `json:"major"`
}

type Heartbeat struct {
	Name       string    `json:"name"`
	Version    *Version  `json:"version"`
	CurrentUTC time.Time `json:"currentutc"`
	Hostname   string    `json:"hostname"`
	GitVersion string    `json:"gitversion"`
}
