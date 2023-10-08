package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func GetLoggingEnv() string {
	checkRunningEnv := os.Getenv("HEX_ARCH_ENV")
	if checkRunningEnv == "release" {
		return "structured"
	}
	return "stdout"
}

func SetupLogger() {
	Log = CreateLoggerInstant()
}

func CreateLoggerInstant() *logrus.Logger {
	logInstance := logrus.New()
	logInstance.SetOutput(io.MultiWriter(os.Stdout))
	logInstance.SetReportCaller(true)

	if GetLoggingEnv() == "structured" {
		logInstance.SetLevel(logrus.ErrorLevel)
		logInstance.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logInstance.SetLevel(logrus.DebugLevel)
		logInstance.SetFormatter(&myFormatter{logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
		}})
	}
	return logInstance
}

type myFormatter struct {
	logrus.TextFormatter
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	strList := strings.Split(entry.Caller.File, "/")
	fileName := strList[len(strList)-1]

	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 31
	case logrus.WarnLevel:
		levelColor = 33
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31
	default:
		levelColor = 36
	}
	return []byte(fmt.Sprintf("[%s] - %s - [line:%d] - \x1b[%dm%s\x1b[0m - %s. Data: %v\n", entry.Time.Format(f.TimestampFormat), fileName, entry.Caller.Line, levelColor,
		strings.ToUpper(entry.Level.String()), entry.Message, entry.Data)), nil
}
