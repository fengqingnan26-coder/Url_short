package database

import (
	"log"

	"github.com/jekyulll/url_shortener/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	switch cfg.Driver {
	case "mysql":
		return newMySQLDB(cfg)
	case "postgre":
	case "mongodb":
	}
	return nil, nil
}

func newMySQLDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DNS()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB instance: %v", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	log.Printf("connect to db: %s:%d", cfg.Host, cfg.Port)
	return db, nil
}
