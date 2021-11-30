package types

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	URL      string         `json:"url"`
	Path     string         `json:"path"`
	Status   int            `json:"status"`
	Commands chan int       `gorm:"-" json:"-"`
	Command  int            `gorm:"-"`
	Progress Progress       `gorm:"-"`
	Callback func(job *Job) `gorm:"-" json:"-"`
}

type HttpURL struct {
	gorm.Model
	URL      string `json:"url"`      // https://acme.com/baseline/update/1/collection.tar.gz/md5 ->
	Filename string `json:"filename"` // /acme.com/baseline_update_1_collection.tar.gz_md5
}

type RepositoryX struct {
	gorm.Model
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Path        string    `json:"path"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Status      int       `json:"status"`
	Interval    int       `json:"interval"`
	LastSuccess time.Time `json:"lastsuccess"`
	LastFail    time.Time `json:"lastfail"`
	Available   bool      `json:"available"`
	Recursive   bool      `json:"recursive"`
}
