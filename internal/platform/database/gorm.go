package database

import (
	"context"
	"fmt"
	"time"

	"github.com/caiohenrique/go-api-template/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open abre conexão GORM com PostgreSQL. O caller deve invocar sqlDB.Close ao encerrar.
func Open(ctx context.Context, cfg *config.Config) (*gorm.DB, func() error, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN()), gormCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("gorm open: %w", err)
	}

	if cfg.GormDebugEnabled() {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("gorm sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	closeFn := func() error {
		return sqlDB.Close()
	}

	return db, closeFn, nil
}
