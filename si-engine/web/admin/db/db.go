package db

import (
	"fmt"
	"log"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase opens (and migrates) the si-engine database
func ConnectDatabase(path string) error {
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database %s: %w", path, err)
	}

	if err := database.AutoMigrate(&User{}, &Log{}, &KeyValuePair{}, &NetCapture{}, &Certificate{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	DB = database

	if err := MigrateUsers(DB); err != nil {
		return fmt.Errorf("failed to migrate users: %w", err)
	}

	return nil
}

// MigrateUsers upgrades user records created by earlier versions:
//   - plain text passwords are replaced by bcrypt hashes
//   - users without a role become viewers
//   - if there is no admin, the oldest user is promoted to admin so that
//     nobody gets locked out after the upgrade
func MigrateUsers(database *gorm.DB) error {
	var users []User
	if err := database.Find(&users).Error; err != nil {
		return err
	}

	for i := range users {
		u := &users[i]
		changed := false

		if u.Password != "" && !IsPasswordHash(u.Password) {
			hash, err := hashWithoutPolicy(u.Password)
			if err != nil {
				return err
			}
			u.Password = hash
			changed = true
			log.Printf("Migrated plain text password to bcrypt for user %s", u.UserName)
		}

		if u.Role != RoleAdmin && u.Role != RoleViewer {
			u.Role = RoleViewer
			changed = true
		}

		if changed {
			if err := database.Save(u).Error; err != nil {
				return err
			}
		}
	}

	var admins int64
	if err := database.Model(&User{}).Where("role = ?", RoleAdmin).Count(&admins).Error; err != nil {
		return err
	}

	if admins == 0 && len(users) > 0 {
		var first User
		if err := database.Order("id").First(&first).Error; err != nil {
			return err
		}
		first.Role = RoleAdmin
		if err := database.Save(&first).Error; err != nil {
			return err
		}
		log.Printf("No administrator found, promoted %s to admin", first.UserName)
	}

	return nil
}
