package flags

import (
	"fmt"
	"strings"
)

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

var AllowedLogLevels = []string{string(Debug), string(Info), string(Warn), string(Error)}

func (f *LogLevel) String() string {
	return string(*f)
}

func (f *LogLevel) Type() string {
	return "LogLevel"
}

func (f *LogLevel) Set(value string) error {
	for _, logLevel := range AllowedLogLevels {
		if logLevel == value {
			*f = LogLevel(value)
			return nil
		}
	}
	return fmt.Errorf("invalid log level. Allowed values: %s", strings.Join(AllowedLogLevels, ", "))
}
