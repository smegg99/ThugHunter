// core/datastore/datastore.go
package datastore

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"smuggr.xyz/thughunter/common/models"
)

var DB *gorm.DB

func Initialize(path string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := DB.AutoMigrate(&models.Host{}); err != nil {
		log.Fatalf("migrate error: %v", err)
	}
}
