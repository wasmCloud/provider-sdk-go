package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Level string

const (
	Error    Level = "error"
	Warn     Level = "warn"
	Info     Level = "info"
	Debug    Level = "debug"
	Trace    Level = "trace"
	Critical Level = "critical"
)

func (l Level) String() string {
	return string(l)
}

func (l *Level) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "error":
		*l = Error
	case "warn":
		*l = Warn
	case "info":
		*l = Info
	case "debug":
		*l = Debug
	case "trace":
		*l = Trace
	case "critical":
		*l = Critical
	default:
		return fmt.Errorf("invalid level: %s", s)
	}

	return nil
}
