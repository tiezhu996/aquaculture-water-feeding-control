package database

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/model"
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(pingCtx)
}

func Open(ctx context.Context, databaseURL, driver, environment string) (*gorm.DB, error) {
	level := logger.Warn
	if environment == "development" {
		level = logger.Info
	}
	var dialector gorm.Dialector
	switch driver {
	case "sqlite":
		dialector = sqlite.Open(databaseURL)
	default:
		dialector = postgres.Open(databaseURL)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:  logger.Default.LogMode(level),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Pond{},
		&model.WaterReading{},
		&model.FeedingPlan{},
		&model.ControlExecution{},
		&model.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := seed(context.Background(), db); err != nil {
		return nil, fmt.Errorf("seed database: %w", err)
	}
	return db, nil
}

func seed(ctx context.Context, db *gorm.DB) error {
	db = db.WithContext(context.Background())
	users := []struct {
		Username, DisplayName, Password string
		Role                            constants.Role
	}{
		{"admin", "系统管理员", "admin123", constants.RoleAdmin},
		{"manager", "生产主管", "manager123", constants.RoleManager},
		{"operator", "值班操作员", "operator123", constants.RoleOperator},
		{"viewer", "观察员", "viewer123", constants.RoleViewer},
	}
	for _, entry := range users {
		var count int64
		if err := db.Model(&model.User{}).Where("username = ?", entry.Username).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(entry.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user := model.User{Username: entry.Username, DisplayName: entry.DisplayName, PasswordHash: string(hash), Role: entry.Role, Active: true}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}
	var pondCount int64
	if err := db.Model(&model.Pond{}).Count(&pondCount).Error; err != nil {
		return err
	}
	if pondCount == 0 {
		ponds := []model.Pond{
			{Code: "P-A01", Name: "东区一号塘", Species: "南美白对虾", AreaSquareMeters: 3600, CapacityKg: 12000, GrowthStage: "成长期", Status: constants.PondStatusActive, Manager: "李海", Notes: "主生产塘"},
			{Code: "P-B03", Name: "西区三号塘", Species: "加州鲈鱼", AreaSquareMeters: 2800, CapacityKg: 8500, GrowthStage: "幼鱼期", Status: constants.PondStatusQuarantine, Manager: "王宁", Notes: "近期氨氮偏高"},
		}
		if err := db.Create(&ponds).Error; err != nil {
			return err
		}
		log.Printf("seeded %d ponds", len(ponds))
	}
	return nil
}
