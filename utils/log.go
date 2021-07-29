package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var customLog = logrus.New()

type Logger struct {
	Instance *logrus.Logger
}

func init() {
	customLog.Out = os.Stdout
	customLog.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func SetLogLevel() {
	level := GetEnv("LOG_LEVEL")
	if level == logrus.DebugLevel.String() {
		customLog.Level = logrus.DebugLevel
	} else {
		customLog.Level = logrus.InfoLevel
	}
}

func LoggerInstance() *logrus.Logger {
	return customLog
}
