package logs

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/ebarthur/jotl/cmd/config"
	"github.com/ebarthur/jotl/cmd/db"
	t "github.com/ebarthur/jotl/cmd/types"
)

type LogStreamer struct {
	db         *db.LogsDB
	logChan    chan t.LogEntry
	stopChan   chan struct{}
	wg         sync.WaitGroup
	stdoutPipe *os.File
	stderrPipe *os.File
	logConfig  *t.LogEntry
	stdoutOld  *os.File
	stderrOld  *os.File
}

func NewLogStreamer(logsDB *db.LogsDB, logConfig *t.LogEntry) *LogStreamer {
	return &LogStreamer{
		db:        logsDB,
		logChan:   make(chan t.LogEntry, 1000), // Buffered channel
		stopChan:  make(chan struct{}),
		logConfig: logConfig,
	}
}

func (ls *LogStreamer) Start() {
	ls.wg.Add(1)
	go func() {
		defer ls.wg.Done()
		for {
			select {
			case <-ls.stopChan:
				return
			case logEntry := <-ls.logChan:
				err := ls.db.Insert(
					logEntry.Level,
					logEntry.Message,
					logEntry.Environment,
					logEntry.ServiceName,
					logEntry.Metadata,
				)
				if err != nil {
					os.Stderr.WriteString("Failed to log to database: " + err.Error() + "\n")
				}
			}
		}
	}()
}

func (ls *LogStreamer) Stop() {
	close(ls.stopChan)
	ls.wg.Wait()
	close(ls.logChan)
	ls.restoreStd()
}

// Log allows external callers to send log entries to the log channel
func (ls *LogStreamer) Log(entry t.LogEntry) {
	// Use logConfig as a template for each log entry
	entry.Environment = ls.logConfig.Environment
	entry.ServiceName = ls.logConfig.ServiceName

	select {
	case ls.logChan <- entry:
		// Successfully added to the channel
	default:
		// Channel is full; This may rarely happen. []: handle overflow
		errorMsg := fmt.Sprintf("Log channel is full. Dropping log entry: %v", entry)
		log.Println(errorMsg)
		os.Stderr.WriteString(errorMsg + "\n")
	}
}

// RedirectStd redirects stdout and stderr to the log channel
func (ls *LogStreamer) RedirectStd(config *config.JotlConfig) error {
	ls.stdoutOld = os.Stdout
	ls.stderrOld = os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v\n", err)
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	ls.stdoutPipe = stdoutW
	os.Stdout = stdoutW

	go ls.captureLogs(stdoutR, string(config.Logging.Level))

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	ls.stderrPipe = stderrW
	os.Stderr = stderrW

	go ls.captureLogs(stderrR, string(config.Logging.Level))

	return nil
}

// restoreStd restores the original stdout and stderr
func (ls *LogStreamer) restoreStd() {
	os.Stdout = ls.stdoutOld
	os.Stderr = ls.stderrOld
	ls.stdoutPipe.Close()
	ls.stderrPipe.Close()
}

// captureLogs captures logs from redirected stdout or stderr and sends them to the log channel
func (ls *LogStreamer) captureLogs(reader io.Reader, configLevel string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logEntry := t.LogEntry{
			Message: scanner.Text(),
		}

		// Determine the log level based on the source
		if configLevel == "info" {
			logEntry.Level = "INFO"
		} else if configLevel == "warn" {
			logEntry.Level = "warn"
		} else if configLevel == "error" {
			logEntry.Level = "ERROR"
		} else if configLevel == "debug" {
			logEntry.Level = "DEBUG"
		}

		// Filter logs based on the config level
		if shouldLog(logEntry.Level, configLevel) {
			ls.Log(logEntry)
		}
	}
}

// shouldLog determines if a log entry should be logged based on the config level
func shouldLog(entryLevel, configLevel string) bool {
	levels := map[string]int{
		"DEBUG": 0,
		"INFO":  1,
		"WARN":  2,
		"ERROR": 3,
	}

	return levels[entryLevel] >= levels[configLevel]
}
