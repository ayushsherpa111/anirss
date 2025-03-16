package dbhandler

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbLocation = "./store/anilookup.db"
)

var tables = []string{
	`CREATE TABLE IF NOT EXISTS ANIME(
      ID    INTEGER,
      NAME  VARCHAR,
      STATUS VARCHAR,
      PRIMARY KEY(ID)
    )`,
	`CREATE TABLE IF NOT EXISTS EPISODES(
      ANI_ID INTEGER,
      SEASON_ID INTEGER,
      EPISODE_NUMBER INTEGER,
      QUALITY VARCHAR,
      FOREIGN KEY(ANI_ID) REFERENCES ANIME(ID),
      PRIMARY KEY(ANI_ID, SEASON_ID, EPISODE_NUMBER)
  )`,
}

func InitDB(dbLogger *slog.Logger) (db *sql.DB) {
	var err error
	if db, err = sql.Open("sqlite3", dbLocation); err != nil {
		dbLogger.Error("failed to connect to database", "err", err.Error())
		os.Exit(1)
	}
	ensureDB(db, dbLogger)
	return
}

func ensureDB(db *sql.DB, dbLogger *slog.Logger) {
	txn, err := db.Begin()
	if err != nil {
		dbLogger.Error("Failed to start transaction", "err", err.Error())
		panic(err)
	}
	for _, table := range tables {
		dbLogger.Info("Creating table.", "sql", table)
		if _, err := txn.Exec(table); err != nil {
			dbLogger.Error("Failed to create table", "err", err.Error())
			txn.Rollback()
		} else {
			dbLogger.Info("Created table.")
		}
	}
	txn.Commit()
}
