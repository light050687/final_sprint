package database

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func CreateDatabase(dbFile string) error {
	file, err := os.Create(dbFile)
	if err != nil {
		return err
	}
	file.Close()

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	createTableQuery := `
 CREATE TABLE IF NOT EXISTS scheduler (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  title TEXT NOT NULL,
  comment TEXT,
  repeat TEXT
 );
 CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
 `

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	log.Println("Таблица scheduler успешно создана")
	return nil
}
