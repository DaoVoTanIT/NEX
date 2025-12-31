package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/create-go-app/fiber-go-template/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GORMMysqlConnection func for connection to MySQL database using GORM.
func GORMMysqlConnection() (*gorm.DB, error) {
	// Define database connection settings.
	maxConn, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	maxIdleConn, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
	maxLifetimeConn, _ := strconv.Atoi(os.Getenv("DB_MAX_LIFETIME_CONNECTIONS"))

	// Build MySQL connection URL.
	mysqlConnURL, err := utils.ConnectionURLBuilder("mysql")
	if err != nil {
		return nil, err
	}

	// Define database connection for MySQL with GORM.
	db, err := gorm.Open(mysql.Open(mysqlConnURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error, not connected to database, %w", err)
	}

	// Get underlying sql.DB to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting underlying sql.DB, %w", err)
	}

	// Set database connection settings:
	// 	- SetMaxOpenConns: the default is 0 (unlimited)
	// 	- SetMaxIdleConns: defaultMaxIdleConns = 2
	// 	- SetConnMaxLifetime: 0, connections are reused forever
	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetimeConn))

	// Try to ping database.
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error, not sent ping to database, %w", err)
	}

	return db, nil
}
