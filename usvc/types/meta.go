package types

import "time"

type MetaOpRequest struct {
	Operation string `json:"operation"`
}

type MetaGetRequest struct {
	Name string `json:"name"`
}

type MetaSetRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MetaLog struct {
	Time        time.Time `json:"time"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}
