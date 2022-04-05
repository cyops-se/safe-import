package types

import (
	"time"

	"gorm.io/gorm"
)

type ListInfo struct {
	Class   string `json:"class"`
	Allowed bool   `json:"allowed"`
	NoScan  bool   `json:"noscan"`
}

type HttpRequest struct {
	gorm.Model
	ListInfo
	Time     time.Time `json:"time"`
	Type     string    `json:"type"`
	FromIP   string    `json:"fromip"`
	ToHost   string    `json:"tohost"`
	Method   string    `json:"method"`
	URL      string    `json:"url"`
	MatchURL string    `json:"matchurl"`
	Headers  string    `json:"headers"`
	LastSeen time.Time `json:"lastseen"`
	Count    uint      `json:"count"`
}

type DnsRequest struct {
	gorm.Model
	ListInfo
	Time       time.Time `json:"time"`
	FromIP     string    `json:"fromip"`
	Query      string    `json:"query"`
	MatchQuery string    `json:"matchquery"`
	LastSeen   time.Time `json:"lastseen"`
	Count      uint      `json:"count"`
}

type Repository struct {
	gorm.Model
	FailureMsg  string    `json:"failuremsg"`
	LastSuccess time.Time `json:"lastsuccess"`
	LastFailure time.Time `json:"lastfailure"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	MatchURL    string    `json:"matchurl"`
	Hash        string    `json:"hash" gorm:"unique"`
	Method      string    `json:"method"`
	Headers     string    `json:"headers"`
	Recursive   bool      `json:"recursive"`
	Host        string    `json:"host"`
	OuterPath   string    `json:"outerpath"`
	InnerPath   string    `json:"innerpath"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Anonymous   bool      `json:"anonymous"`
	Available   bool      `json:"available"`
}
