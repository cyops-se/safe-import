package common

import (
	"fmt"

	"github.com/cyops-se/safe-import/si-inner/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("si-inner.db"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	// dsn := "user=dev password=hemligt dbname=dev host=localhost port=5432"
	// database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to database", err)
		return
	}

	fmt.Println("Database connected!")

	database.AutoMigrate(&types.HttpRequest{}, &types.DnsRequest{}, &types.Repository{})

	DB = database
}
