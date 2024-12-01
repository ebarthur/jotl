package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	t "github.com/ebarthur/jotl/cmd/types"
	_ "github.com/lib/pq"
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
	var createTableQuery string

	if strings.HasPrefix(l.dbpath, "postgres://") || strings.HasPrefix(l.dbpath, "postgresql://") {
		createTableQuery = `CREATE TABLE IF NOT EXISTS logs (
					id SERIAL PRIMARY KEY,
					level TEXT NOT NULL,
					environment TEXT,
					service_name TEXT,
					message TEXT NOT NULL,
					metadata TEXT,
					timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`
	} else {
		createTableQuery = `CREATE TABLE IF NOT EXISTS logs (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					level TEXT NOT NULL,
					environment TEXT,
					service_name TEXT,
					message TEXT NOT NULL,
					metadata TEXT,
					timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
			)`
	}

	_, err := l.db.Exec(createTableQuery)
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
	var db *sql.DB
	var err error

	log.Printf("Creating new LogsDB with dbpath: %s", dbpath)

	if strings.HasPrefix(dbpath, "postgres://") || strings.HasPrefix(dbpath, "postgresql://") {
		log.Println("Detected PostgreSQL database")
		db, err = sql.Open("postgres", dbpath)
		if err != nil {
			log.Printf("Failed to open PostgreSQL database: %v", err)
			return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
		}
	} else {
		log.Println("Detected SQLite database")
		if err := InitLogsDir(filepath.Dir(dbpath)); err != nil {
			log.Printf("Failed to create directory for logs database: %v", err)
			return nil, fmt.Errorf("failed to create directory for logs database: %w", err)
		}

		db, err = sql.Open("sqlite3", dbpath)
		if err != nil {
			log.Printf("Failed to open SQLite database: %v", err)
			return nil, fmt.Errorf("failed to open SQLite database: %w", err)
		}
	}

	if err := db.Ping(); err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to the database")

	logsDB := &LogsDB{
		db:     db,
		dbpath: dbpath,
	}

	if err := logsDB.CreateTable(); err != nil {
		log.Printf("Failed to create logs table: %v", err)
		return nil, fmt.Errorf("failed to create logs table: %w", err)
	}

	log.Println("Successfully created logs table")

	return logsDB, nil
}

func ConnectDatabase(connStr string) (*sql.DB, error) {
	var driverName string

	if strings.HasPrefix(connStr, "postgres://") {
		driverName = "postgres"
	} else if strings.HasPrefix(connStr, "file://") {
		driverName = "sqlite3"
		connStr = strings.TrimPrefix(connStr, "file://")
	} else {
		return nil, fmt.Errorf("unsupported database type")
	}

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return db, nil
}
