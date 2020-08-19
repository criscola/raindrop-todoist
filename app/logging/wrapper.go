package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

// StandardLogger enforces specific log message formats
type StandardLogger struct {
	*zerolog.Logger
}

// New initializes the standard logger
func New() *StandardLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	prettyOutput := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	logger := zerolog.New(prettyOutput).With().Timestamp().Logger()
	return &StandardLogger{&logger}
}
