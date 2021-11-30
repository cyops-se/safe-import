package types

import "time"

type ByIdRequest struct {
	ID int `json:"id"`
}

type NameValue struct {
	Name  string `json:"Name"`
	Value string `json:"value"`
}

type WaitRequest struct {
	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Headers []NameValue `json:"headers"`
	Body    string      `json:"body"`
	NoScan  bool        `json:"noscan"`
}

type Response struct {
	Success bool   `json:"success"`
	Error   error  `json:"error"`
	Message string `json:"message"`
}

type WaitResponse struct {
	Success  bool        `json:"success"`
	Error    error       `json:"error"`
	Filename string      `json:"filename"`
	Headers  []NameValue `json:"headers"`
}

type Total struct {
	Percent float64 `json:"percent"`
	Size    int64   `json:"size"`
	Total   int64   `json:"total"`
}

type Progress struct {
	ID           int    `json:"repositoryid"`
	Error        error  `json:"-"`
	ErrorMessage string `json:"error"`
	Current      Total  `json:"current"`
	Total        Total  `json:"total"`
	CurrentPath  string `json:"currentpath"`
}

type Repository struct {
	ID          int       `json:"id"`
	FailureMsg  string    `json:"failuremsg"`
	LastSuccess time.Time `json:"lastsuccess"`
	LastFailure time.Time `json:"lastfailure"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	MatchURL    string    `json:"matchurl"`
	Hash        string    `json:"hash"`
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
