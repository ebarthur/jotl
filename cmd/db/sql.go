package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	t "github.com/ebarthur/jotl/cmd/types"
	_ "github.com/mattn/go-sqlite3"
)

type LogsDB struct {
	db     *sql.DB
	dbpath string
}

func InitLogsDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0o770)
		}
		return err
	}
	return nil
}

func (l *LogsDB) TableExists(name string) bool {
	if _, err := l.db.Query("SELECT * FROM logs"); err == nil {
		return true
	}
	return false
}

func (l *LogsDB) CreateTable() error {
	_, err := l.db.Exec(`CREATE TABLE IF NOT EXISTS logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        level TEXT NOT NULL,
        environment TEXT,
        service_name TEXT,
        message TEXT NOT NULL,
        metadata TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
	return err
}

func (l *LogsDB) Insert(level, message, environment, serviceName, metadata string) error {
	_, err := l.db.Exec(
		`INSERT INTO logs(level, message, environment, service_name, metadata) 
         VALUES(?, ?, ?, ?, ?)`,
		level, message, environment, serviceName, metadata,
	)
	return err
}

func (l *LogsDB) SaveLogEntry(logEntry t.LogEntry) error {
	metadata, err := json.Marshal(logEntry.Metadata)
	if err != nil {
		return fmt.Errorf("unable to marshal metadata: %w", err)
	}

	return l.Insert(
		logEntry.Level,
		logEntry.Message,
		logEntry.Environment,
		logEntry.ServiceName,
		string(metadata),
	)
}

func (l *LogsDB) GetLogs() ([]t.LogEntry, error) {
	var logs []t.LogEntry
	rows, err := l.db.Query("SELECT id, level, message, environment, service_name, metadata, timestamp FROM logs")
	if err != nil {
		return logs, fmt.Errorf("unable to get values: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var log t.LogEntry
		err = rows.Scan(
			&log.ID,
			&log.Level,
			&log.Message,
			&log.Environment,
			&log.ServiceName,
			&log.Metadata,
			&log.Timestamp,
		)
		if err != nil {
			return logs, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func NewLogsDB(dbpath string) (*LogsDB, error) {
	// Ensure the logs directory exists
	if err := InitLogsDir(filepath.Dir(dbpath)); err != nil {
		return nil, fmt.Errorf("failed to create directory for logs database: %w", err)
	}

	// Open the SQLite database file
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Verify connection by pinging the database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	logsDB := &LogsDB{
		db:     db,
		dbpath: dbpath,
	}

	// Attempt to create the logs table
	if err := logsDB.CreateTable(); err != nil {
		return nil, fmt.Errorf("failed to create logs table: %w", err)
	}

	return logsDB, nil
}
