package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
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
	logConfig  *t.LogEntry
	metadataRe *regexp.Regexp
}

func NewLogStreamer(logsDB *db.LogsDB, logConfig *t.LogEntry) *LogStreamer {
	return &LogStreamer{
		db:         logsDB,
		logChan:    make(chan t.LogEntry, 1000), // Buffered channel
		stopChan:   make(chan struct{}),
		logConfig:  logConfig,
		metadataRe: regexp.MustCompile(`\{.*\}`), // Regex to extract JSON-like structures
	}
}

// Enhanced log entry parsing function
func (ls *LogStreamer) parseLogEntry(message string) t.LogEntry {
	// ANSI color escape sequence regex
	var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*[mz]`)

	// Strip ANSI color codes
	cleanMessage := ansiColorRegex.ReplaceAllString(message, "")

	// Trim whitespace
	cleanMessage = strings.TrimSpace(cleanMessage)

	// Extract potential metadata
	metadataMatch := ls.metadataRe.FindString(cleanMessage)

	// Remove metadata from the original message
	cleanMessage = ls.metadataRe.ReplaceAllString(cleanMessage, "")

	// Initialize log entry
	logEntry := t.LogEntry{
		Message:  cleanMessage,
		Level:    "", // Determine dynamically
		Metadata: make(map[string]interface{}),
	}

	// Try to parse metadata as JSON
	if metadataMatch != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataMatch), &metadata); err == nil {
			logEntry.Metadata = metadata
		}
	}

	// Determine log level based on known prefixes
	levels := []struct {
		prefix string
		level  string
	}{
		{"DEBUG:", "DEBUG"},
		{"INFO:", "INFO"},
		{"WARN:", "WARN"},
		{"ERROR:", "ERROR"},
	}

	for _, lvl := range levels {
		if strings.HasPrefix(cleanMessage, lvl.prefix) {
			logEntry.Level = lvl.level
			logEntry.Message = strings.TrimPrefix(cleanMessage, lvl.prefix) // Remove the prefix
			break
		}
	}

	// Fallback: Use user's configured level if none detected
	if logEntry.Level == "" {
		logEntry.Level = ls.logConfig.Level
	}

	return logEntry
}

// Start begins streaming logs
func (ls *LogStreamer) Start() {
	ls.wg.Add(1)
	go func() {
		defer ls.wg.Done()
		for {
			select {
			case logEntry := <-ls.logChan:
				// Process log entry
				err := ls.db.SaveLogEntry(logEntry)
				if err != nil {
					log.Printf("Failed to save log entry: %v", err)
				}
			case <-ls.stopChan:
				return
			}
		}
	}()
}

// Stop stops streaming logs
func (ls *LogStreamer) Stop() {
	close(ls.stopChan)
	ls.wg.Wait()
	close(ls.logChan)
}

func (ls *LogStreamer) Log(entry t.LogEntry) {
	entry.Environment = ls.logConfig.Environment
	entry.ServiceName = ls.logConfig.ServiceName
	select {
	case ls.logChan <- entry:
		// Add log to channel
	default:
		// Channel is full
		errorMsg := fmt.Sprintf("Log channel is full. Dropping log entry: %v", entry)
		log.Println(errorMsg)
		os.Stderr.WriteString(errorMsg + "\n")
	}
}

func (ls *LogStreamer) RedirectStd(config *config.JotlConfig) error {
	// Create pipes for stdout and stderr
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Redirect system stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Start goroutines to process logs
	go ls.captureAndTeeLogs(stdoutReader, origStdout, string(config.Logging.Level))
	go ls.captureAndTeeLogs(stderrReader, origStderr, string(config.Logging.Level))

	return nil
}

func (ls *LogStreamer) captureAndTeeLogs(reader io.Reader, originalOutput *os.File, configLevel string) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		message := scanner.Text()

		// Forward the log to the original terminal
		fmt.Fprintln(originalOutput, message)

		// Parse the log entry
		logEntry := ls.parseLogEntry(message)

		// Filter logs based on the user's configuration
		if shouldLog(logEntry.Level, configLevel) {
			ls.Log(logEntry)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs: %v", err)
	}
}

func shouldLog(entryLevel, configLevel string) bool {
	levels := map[string]int{
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  3,
		"ERROR": 4,
	}

	return levels[entryLevel] >= levels[configLevel]
}
