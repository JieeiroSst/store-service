package log

import (
	"log/syslog"
	"os"
	"runtime"

	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	airbrake "gopkg.in/gemnasium/logrus-airbrake-hook.v2" // the package is named "airbrake"
)

// var logger = &logrus.Logger{
// 	Out:   os.Stderr,
// 	Level: logrus.DebugLevel,
// 	Formatter: &easy.Formatter{
// 		TimestampFormat: "2006-01-02 15:04:05",
// 		LogFormat:       "[%lvl%]: %time% - %msg%",
// 	},
// }

func init() {
	log.AddHook(airbrake.NewHook(123, "xyz", "production"))
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

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

func logger() *log.Entry {
	var logger *log.Entry
	if pc, file, line, ok := runtime.Caller(1); ok {
		fName := runtime.FuncForPC(pc).Name()
		return logger.WithField("file", file).WithField("line", line).WithField("func", fName)
	}
	return logger
}

func Trace(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Trace(string(msgJson))
}

func Debug(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Debugf("%v \n", string(msgJson))
}

func Info(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Infof("%v \n", string(msgJson))
}

func Warn(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Warnf("%v \n", string(msgJson))
}

func Error(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Errorf("%v \n", string(msgJson))
}

func Fatal(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Fatalf("%v \n", string(msgJson))
}

func Panic(msg interface{}) {
	msgJson, _ := json.Marshal(&msg)

	logger().Panicf("%v \n", string(msgJson))
}
