package db

import (
	"database/sql"
	"os"
)

func RunMigrations(database *sql.DB) error {
	query, err := os.ReadFile(".internal/db/migrations")
	if err != nil {
		return err
	}

	_, err = database.Exec(string(query))
	return err
}
