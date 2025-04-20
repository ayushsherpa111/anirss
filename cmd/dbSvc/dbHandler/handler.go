package dbhandler

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ayushsherpa111/anirss/pkg/objects"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbLocation = "./store/anilookup.db"
)

var tables = []string{
	`CREATE TABLE IF NOT EXISTS ANIME(
      ID        INTEGER,
      TITLE     VARCHAR,
      START_DATE DATE,
      END_DATE DATE,
      STATUS VARCHAR,
      PRIMARY KEY(ID)
    )`,
	`CREATE TABLE IF NOT EXISTS EPISODES(
      ANI_ID INTEGER,
      EP_NUM INTEGER,
      TITLE VARCHAR,
      AIR_DATE DATE,
      FOREIGN KEY(ANI_ID) REFERENCES ANIME(ID),
      PRIMARY KEY(ANI_ID, EP_NUM)
  )`,
	`CREATE TABLE IF NOT EXISTS DOWNLOADS(
    ANI_ID  INTEGER,
    EP_NUM  INTEGER,
    QUALITY VARCHAR,
    MAGNET  VARCHAR,
    STATUS  VARCHAR,
    PRIMARY KEY (ANI_ID, EP_NUM),
    FOREIGN KEY(ANI_ID, EP_NUM) REFERENCES EPISODES(ANI_ID, EP_NUM)
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

func insert(db *sql.DB, tableName string, rows []string, cols []any) (int64, error) {
	txn, err := db.Begin()
	if err != nil {
		return 0, err
	}
	qMarks := make([]string, 0, len(rows))
	for range len(rows) {
		qMarks = append(qMarks, "?")
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) Values(%s)",
		tableName, strings.Join(rows, ","), strings.Join(qMarks, ", "))
	fmt.Println(query)
	stmt, err := txn.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(cols...)
	if err != nil {
		return 0, err
	}
	txn.Commit()
	return result.RowsAffected()
}

func insertObj(db *sql.DB, done <-chan bool, dbChan <-chan objects.DBRecords, logChan chan<- objects.Logging) chan int {
	dbResChan := make(chan int)
	go func() {
		defer close(dbResChan)
		for {
			select {
			case <-done:
				return
			case obj, ok := <-dbChan:
				if !ok {
					return
				}
				columns, values := obj.GetDBRecords()
				rows, err := insert(db, obj.GetTblName(), columns, values)
				if err != nil {
					logChan <- objects.Logging{
						Message: "failed to insert row",
						Error:   err,
						Level:   objects.L_ERROR,
					}
				} else {
					logChan <- objects.Logging{
						Message: fmt.Sprintf("Insert rows %d", rows),
						Error:   nil,
						Level:   objects.L_INFO,
					}
				}
				dbResChan <- int(rows)
			}
		}
	}()
	return dbResChan
}
