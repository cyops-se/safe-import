package db

import (
	"path/filepath"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("a long enough password")
	if err != nil {
		t.Fatal(err)
	}
	if !IsPasswordHash(hash) {
		t.Errorf("IsPasswordHash(%q) = false", hash)
	}
	if !CheckPassword(hash, "a long enough password") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "wrong password") {
		t.Error("wrong password accepted")
	}
	if _, err := HashPassword("short"); err != ErrPasswordTooShort {
		t.Errorf("short password: expected ErrPasswordTooShort, got %v", err)
	}
}

func TestMigrateUsers(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "m.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := database.DB(); err == nil {
			sqlDB.Close()
		}
	})
	if err := database.AutoMigrate(&User{}); err != nil {
		t.Fatal(err)
	}

	// Users as created by earlier versions: plain text passwords and no role
	database.Create(&User{UserName: "first@example.com", Password: "secret", FullName: "First", Role: ""})
	database.Create(&User{UserName: "second@example.com", Password: "secret2", FullName: "Second", Role: ""})

	if err := MigrateUsers(database); err != nil {
		t.Fatal(err)
	}

	var users []User
	database.Order("id").Find(&users)
	for _, u := range users {
		if !IsPasswordHash(u.Password) {
			t.Errorf("%s: password not hashed: %q", u.UserName, u.Password)
		}
	}
	if !CheckPassword(users[0].Password, "secret") {
		t.Error("migrated password does not match the original")
	}
	if users[0].Role != RoleAdmin {
		t.Errorf("oldest user should be promoted to admin, got %q", users[0].Role)
	}
	if users[1].Role != RoleViewer {
		t.Errorf("other users should be viewers, got %q", users[1].Role)
	}

	// Running the migration again must not change anything
	before := users[0].Password
	if err := MigrateUsers(database); err != nil {
		t.Fatal(err)
	}
	var again User
	database.First(&again, users[0].ID)
	if again.Password != before {
		t.Error("migration is not idempotent")
	}
}
