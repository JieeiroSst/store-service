package log

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

var Log Logger

func Init(level string, output string) error {
	// Set the log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}
	var err error

	if output == "stdout" && os.Getenv("ENV") != "production" {
		Log.logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false}).With().Timestamp().Logger()
	} else {
		if output == "stdout" {
			Log.logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		} else if output == "stderr" {
			Log.logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		} else {
			f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
				return err
			}
			Log.logger = zerolog.New(f).With().Timestamp().Logger()
		}
	}

	return err 
}

func Debug(msg string) {
    Log.logger.Debug().Msg(msg)
}

func Info(msg string) {
    Log.logger.Info().Msg(msg)
}

func Infof(format string, v ...interface{}) {
    Log.logger.Info().Msgf(format, v...)
}
 
func Warn(msg string) {
    Log.logger.Warn().Msg(msg)
}

func Error(msg string, err error) {
    Log.logger.Err(err).Msg(msg)
}

func Fatal(msg string, err error) {
    Log.logger.Fatal().Err(err).Msg(msg)
}

func Panic(msg string, err error) {
    Log.logger.Panic().Err(err).Msg(msg)
}

