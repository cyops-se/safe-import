package common

import (
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("si-gatekeeper.db"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	// dsn := "user=dev password=hemligt dbname=dev host=localhost port=5432"
	// database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// fmt.Println("Failed to connect to database", err)
		return
	}

	// fmt.Println("Database connected!")

	database.AutoMigrate(&types.CachedItem{})

	DB = database
}
