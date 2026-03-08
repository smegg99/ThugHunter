package datastore

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

var (
	db       *gorm.DB
	initOnce sync.Once
)

func allModels() []interface{} {
	return []interface{}{
		&models.Host{},
		&models.Account{},
		&models.VNCService{},
		&models.RDPService{},
		&models.SPICEService{},
		&models.PKCameraService{},
	}
}

func Initialize(path string) error {
	var initErr error
	initOnce.Do(func() {
		logger.Info().Str("path", path).Msg("initializing datastore")

		logger.Debug().Str("path", path).Msg("opening sqlite database")

		gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})
		if err != nil {
			initErr = fmt.Errorf("open sqlite %s: %w", path, err)
			return
		}

		logger.Debug().Msg("enabling WAL journal mode")

		if err := gormDB.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
			initErr = fmt.Errorf("enable WAL: %w", err)
			return
		}

		logger.Debug().Msg("running auto-migration")

		if err := gormDB.AutoMigrate(allModels()...); err != nil {
			initErr = fmt.Errorf("auto-migrate: %w", err)
			return
		}

		db = gormDB
		logger.Info().Msg("datastore ready")
	})
	return initErr
}

func Get() *gorm.DB {
	if db == nil {
		panic("datastore: not initialised  call Initialize first")
	}
	return db
}
