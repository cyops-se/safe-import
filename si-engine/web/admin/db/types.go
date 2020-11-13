package db

import (
	"time"

	"gorm.io/gorm"
)

type KeyValuePair struct {
	gorm.Model
	Key   string `json:"key"`
	Value string `json:"value"`
	Extra string `json:"extra"`
}

// User
type User struct {
	gorm.Model
	UserName string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password"`
	FullName string `form:"fullname" json:"fullname" binding:"required"`
}

type UserData struct {
	ID       uint   `form:"id" json:"id" binding:"required" gorm:"primary_key"`
	UserName string `form:"email" json:"email" binding:"required"`
	FullName string `form:"fullname" json:"fullname" binding:"required"`
}

type UserPasswordUpdate struct {
	ID       uint   `form:"id" json:"id" binding:"required" gorm:"primary_key"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserCredentials struct {
	UserName string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type Log struct {
	gorm.Model
	Time        time.Time `json:"time"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type NetCapture struct {
	gorm.Model
	Time       time.Time `json:"time"`
	Type       string    `json:"type"`
	FromIP     string    `json:"fromip"`
	ToHost     string    `json:"tohost"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	RequestURI string    `json:"requesturi"`
	Headers    string    `json:"headers"`
	Query      string    `json:"query"`
	Data       string    `json:"data"`
}

type NetRepos struct {
	gorm.Model
	Type      string    `json:"type"`
	ToHost    string    `json:"tohost"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Headers   string    `json:"headers"`
	LocalPath string    `json:"localpath"`
	LastCheck time.Time `json:"lastcheck"`
	LastSync  time.Time `json:"lastsync"`
	State     string    `json:"state"`
}

// type HttpRequestData struct {
// 	Host    string      `json:"host"`
// 	Url     string      `json:"Url"`
// 	Headers http.Header `json:"headers"`
// 	Body    string      `json:"body"`
// }

// type NetHttpCapture struct {
// 	NetCapture
// 	Host    string `json:"host"`
// 	Url     string `json:"url"`
// 	Headers string `json:"headers"`
// 	Body    string `json:"body"`
// }

type Certificate struct {
	gorm.Model
	Name        string `form:"name" json:"name" binding:"required"`
	Description string `form:"description" json:"description" binding:"required"`
}
