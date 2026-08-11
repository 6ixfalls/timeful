package db

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"schej.it/server/models"
)

var orm *gorm.DB

// ORM exposes the configured GORM session for repository code and transactions.
func ORM() *gorm.DB { return orm }

func Init() func() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required")
	}

	var err error
	orm, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		panic(fmt.Errorf("connect to PostgreSQL: %w", err))
	}

	if err := orm.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.EventResponse{},
		&models.Attendee{},
		&models.Folder{},
		&models.FolderEvent{},
		&models.DailyUserLog{},
		&models.FriendRequest{},
		&models.OtpCode{},
	); err != nil {
		panic(fmt.Errorf("migrate PostgreSQL schema: %w", err))
	}
	if err := orm.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email))").Error; err != nil {
		panic(fmt.Errorf("create case-insensitive user email index: %w", err))
	}
	if err := orm.Where("expires_at < ?", time.Now().UTC()).Delete(&models.OtpCode{}).Error; err != nil {
		panic(fmt.Errorf("remove expired OTP codes: %w", err))
	}

	sqlDB, err := orm.DB()
	if err != nil {
		panic(fmt.Errorf("access PostgreSQL connection pool: %w", err))
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return func() {
		if err := sqlDB.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close PostgreSQL connection: %v\n", err)
		}
	}
}
