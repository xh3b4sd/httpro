package logger

import (
	"log"
	"os"

	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

type Config struct {
	Name  string
	Level string
}

// NewLogger creates a new logger with a default backend logging to
// os.Stderr. See github.com/op/go-logging.
func NewLogger(c Config) *logging.Logger {
	// Create new logger.
	logger := logging.MustGetLogger(c.Name)

	// Format logger.
	logging.SetFormatter(logging.MustStringFormatter("[%{level}] %{message}"))

	if c.Level == "" {
		logBackend := logging.NewMemoryBackend(1)
		logging.SetBackend(logBackend)
	} else {
		logBackend := logging.NewLogBackend(os.Stderr, "", log.LstdFlags|log.Lshortfile)
		logBackend.Color = true
		logging.SetBackend(logBackend)
	}

	// Set log level.
	if c.Level != "" {
		logLevel, err := logging.LogLevel(c.Level)
		if err != nil {
			panic(errgo.Mask(err))
		}

		logging.SetLevel(logLevel, c.Name)
	}

	return logger
}
