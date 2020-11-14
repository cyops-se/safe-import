package common

import (
	"fmt"

	"github.com/cyops-se/safe-import/si-outer/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("si-outer.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		fmt.Println("Failed to connect to database", err)
		return
	}

	fmt.Println("Database connected!")

	database.AutoMigrate(&types.Job{})

	DB = database
}
