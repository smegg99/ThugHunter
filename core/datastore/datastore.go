// core/datastore/datastore.go
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
	dbPath   string
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

		// Give concurrent writers time to retry instead of failing immediately.
		if err := gormDB.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
			initErr = fmt.Errorf("set busy_timeout: %w", err)
			return
		}

		// Limit to a single open connection. SQLite does not support truly concurrent writes.
		sqlDB, err := gormDB.DB()
		if err != nil {
			initErr = fmt.Errorf("get sql.DB: %w", err)
			return
		}
		sqlDB.SetMaxOpenConns(1)

		logger.Debug().Msg("running auto-migration")

		if err := gormDB.AutoMigrate(allModels()...); err != nil {
			initErr = fmt.Errorf("auto-migrate: %w", err)
			return
		}

		dbPath = path
		db = gormDB
		logger.Info().Msg("datastore ready")
	})
	return initErr
}

// Get returns the shared GORM connection. Panics if not initialised.
func Get() *gorm.DB {
	if db == nil {
		panic("datastore: not initialised call Initialize first")
	}
	return db
}

// OpenWriter opens a dedicated GORM connection for bulk write operations.
// This bypasses the shared MaxOpenConns(1) pool so that bulk writes do not
// block frontend reads. The caller must close it when done.
// Safe with WAL mode: SQLite allows concurrent readers + 1 writer.
func OpenWriter() (*gorm.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("datastore: not initialised")
	}
	wDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open writer connection: %w", err)
	}
	if err := wDB.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return nil, fmt.Errorf("writer WAL: %w", err)
	}
	if err := wDB.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
		return nil, fmt.Errorf("writer busy_timeout: %w", err)
	}
	sqlDB, err := wDB.DB()
	if err != nil {
		return nil, fmt.Errorf("writer sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	return wDB, nil
}
