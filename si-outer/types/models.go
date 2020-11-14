package types

import "gorm.io/gorm"

type Job struct {
	gorm.Model
	URL    string `json:"url"`
	Path   string `json:"path"`
	Status int    `json:"status"`
}
