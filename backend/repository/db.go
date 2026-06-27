package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database handle.
var DB *gorm.DB

// Init opens the SQLite database and runs AutoMigrate for all models.
func Init(dbPath string) (*gorm.DB, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and foreign keys
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying db: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite is single-writer

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	DB = db
	return db, nil
}

// autoMigrate runs GORM AutoMigrate for all registered models.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Environment{},
		&models.WorkflowExecution{},
		&models.WorkflowStep{},
		&models.Event{},
		&models.AuditLog{},
		&models.Resource{},
		&models.Configuration{},
		&models.Template{},
		&models.Schedule{},
	)
}
