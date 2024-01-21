package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func LoggerInfo(msg string) {
	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()

	logger.Info().Msg(msg)
}

func LoggerError(msg string) {
	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()

	logger.Error().Msg(msg)
}
