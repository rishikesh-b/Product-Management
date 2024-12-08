package logging

import (
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init() {
	// Set formatter based on environment
	env := os.Getenv("APP_ENV")
	if env == "production" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		Logger.SetLevel(logrus.DebugLevel)
	}

	// Enable caller reporting
	Logger.SetReportCaller(true)

	// Log rotation setup
	Logger.SetOutput(&lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     30, // Days
		Compress:   true,
	})

	Logger.Info("Logging initialized")
}
