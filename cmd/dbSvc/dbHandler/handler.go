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
	dbLocation    = "./store/anilookup.db"
	DOWNLOAD_SEED = "seed_downloads"
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

var queryMap = map[string]string{
	// seed the Downloads table with anime ID and Episode ID upon initial download of anime and episodes.
	// status is set to pending to indicate RSS feed to pull magnet URI
	// status will be set downloading once torrent download has started
	// status wil be set to downloaded once torrent succeds.
	"seed_downloads": `
	 INSERT INTO DOWNLOADS
	   (
	     ANI_ID,
	     EP_NUM,
	     STATUS
	   )
	 SELECT
	     a.ID      AS ANI_ID,
	     e.EP_NUM  AS EP_NUM,
	     'PENDING' AS STATUS
	 FROM ANIME a
	 JOIN EPISODES e
	   on e.ANI_ID = a.ID
	 WHERE a.ID = ?
	 `,
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

func buildInsertQuery(tbl_name string, columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return fmt.Sprintf(`INSERT INTO %s( %s ) VALUES(%s?)`,
		tbl_name, strings.Join(columns, ","), strings.Repeat("?,", len(columns)-1))
}

func insertObj(db *sql.DB, done <-chan bool, dataChan <-chan objects.DBRecords, logChan chan<- objects.Logging) chan int {
	dbResChan := make(chan int)
	go func() {
		defer close(dbResChan)
		for {
			select {
			case <-done:
				return
			case obj, ok := <-dataChan:
				if !ok {
					return
				}
				columns, values := obj.GetDBRecords()
				query := buildInsertQuery(obj.GetTblName(), columns)
				rows := insert(db, logChan, query, values...)
				dbResChan <- int(rows)
			}
		}
	}()
	return dbResChan
}

func insert(db *sql.DB, logChan chan<- objects.Logging, query string, args ...any) int32 {
	var err error
	var txn *sql.Tx
	defer func() {
		logChan <- objects.Logging{
			Message: fmt.Sprintf("executing Insert statement, with args: %v", args),
			Payload: fmt.Sprint(query),
			Level:   objects.L_INFO,
		}
		if err != nil {
			logChan <- objects.Logging{
				Message: "error while executing insert statement",
				Level:   objects.L_ERROR,
				Error:   err,
			}
			if txn != nil {
				txn.Rollback()
			}
			return
		}
		txn.Commit()
	}()

	txn, err = db.Begin()
	if err != nil {
		return 0
	}

	stmt, err := txn.Prepare(query)
	if err != nil {
		return 0
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return 0
	}
	rows, _ := res.RowsAffected()

	return int32(rows)
}
