package infrastructure

import "github.com/sirupsen/logrus"

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}
