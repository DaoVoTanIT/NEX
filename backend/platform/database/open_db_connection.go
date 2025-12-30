package database

import (
	"os"
	"sync"

	"github.com/create-go-app/fiber-go-template/app/queries"
	"gorm.io/gorm"
)

// Queries struct for collect all app queries.
type Queries struct {
	*queries.UserQueries        // load queries from User model
	*queries.TaskQueries        // load queries from Task model
	*queries.TaskHistoryQueries // load queries from TaskHistory model
}

var (
	queriesOnce      sync.Once
	queriesInstance  *Queries
	queriesInitError error
)

// OpenDBConnection func for opening database connection with GORM for TaskQueries.
func OpenDBConnection() (*Queries, error) {
	queriesOnce.Do(func() {
		var (
			gormDB *gorm.DB
			err    error
		)

		dbType := os.Getenv("DB_TYPE")

		switch dbType {
		case "pgx":
			_, err = PostgreSQLConnection()
			if err == nil {
				gormDB, err = GORMPostgreSQLConnection()
			}
		case "mysql":
			_, err = MysqlConnection()
			if err == nil {
				gormDB, err = GORMMysqlConnection()
			}
		}

		if err != nil {
			queriesInitError = err
			return
		}

		queriesInstance = &Queries{
			UserQueries:        &queries.UserQueries{DB: gormDB},
			TaskQueries:        &queries.TaskQueries{DB: gormDB},
			TaskHistoryQueries: &queries.TaskHistoryQueries{DB: gormDB},
		}
	})

	if queriesInitError != nil {
		return nil, queriesInitError
	}
	return queriesInstance, nil
}

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
