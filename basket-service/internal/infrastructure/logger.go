package infrastructure

import (
	"log"

	"github.com/sirupsen/logrus"
)

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	log.SetOutput(logrus.StandardLogger().Writer())
	log.SetFlags(0)
}
