package log

import (
	"log/syslog"
	"os"

	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	airbrake "gopkg.in/gemnasium/logrus-airbrake-hook.v2" // the package is named "airbrake"
)

var logger = &logrus.Logger{
	Out:   os.Stderr,
	Level: logrus.DebugLevel,
	Formatter: &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%",
	},
}

func init() {
	log.AddHook(airbrake.NewHook(123, "xyz", "production"))

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	hook, err := logrus_syslog.NewSyslogHook("udp", "localhost:514", syslog.LOG_INFO, "")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon")
	} else {
		log.AddHook(hook)
	}
}

func Trace(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Trace(msgJson)
}

func Debug(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Debug(msgJson)
}

func Info(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Info(msgJson)
}

func Warn(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Warn(msgJson)
}

func Error(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Error(msgJson)
}

func Fatal(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Fatal(msgJson)
}

func Panic(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger.Panic(msgJson)
}
