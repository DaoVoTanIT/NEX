package database

import (
	"os"

	"gorm.io/gorm"
)

// OpenGORMDBConnection func for opening GORM database connection.
func OpenGORMDBConnection() (*gorm.DB, error) {
	// Get DB_TYPE value from .env file.
	dbType := os.Getenv("DB_TYPE")

	// Define a new GORM Database connection with right DB type.
	switch dbType {
	case "pgx":
		return GORMPostgreSQLConnection()
	case "mysql":
		return GORMMysqlConnection()
	default:
		return GORMPostgreSQLConnection() // default to PostgreSQL
	}
}
