package types

import "time"

type LogEntry struct {
	ID          uint
	Level       string                 // INFO, WARN, ERROR, DEBUG
	Message     string                 // main log message
	Environment string                 // node, go
	ServiceName string                 // Config - APP_NAME
	Metadata    map[string]interface{} // JSON or arbitrary text metadata
	Timestamp   time.Time              // Timestamp of the log
}
