package admin

import (
	"log"
	"time"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
)

func Log(category string, title string, msg string) {
	entry := &db.Log{Time: time.Now().UTC(), Category: category, Title: title, Description: msg}
	db.DB.Create(&entry)
	log.Printf("%s: %s, %s", category, title, msg)
	// fmt.Printf("%s: %s, %s\n", category, title, msg)
}
