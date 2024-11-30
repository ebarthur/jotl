package flags

import (
	"fmt"
	"strconv"
)

type Port string

const (
	MinPort = 3000
	MaxPort = 65535
)

func (p Port) String() string {
	return string(p)
}

func (p *Port) Type() string {
	return "Port"
}

func (p *Port) Set(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid port number: must be an integer")
	}

	if port < MinPort || port > MaxPort {
		return fmt.Errorf("port must be between %d and %d", MinPort, MaxPort)
	}

	*p = Port(value)
	return nil
}
