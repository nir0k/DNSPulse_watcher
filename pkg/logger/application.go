package logger

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger


func LogSetup(path string, maxSize int, maxFiles int, maxAge int, severity string) {
	Logger = logrus.New()
	Logger.SetOutput(&lumberjack.Logger{
        Filename: 	path,
		MaxSize:    maxSize,
		MaxBackups: maxFiles,
		MaxAge:     maxAge,
		Compress:   true,
    })

	var notifyLevel logrus.Level
	switch severity {
		case "debug": notifyLevel = logrus.DebugLevel
		case "info": notifyLevel = logrus.InfoLevel
		case "warning": notifyLevel = logrus.WarnLevel
		case "error": notifyLevel = logrus.ErrorLevel
		case "fatal": notifyLevel = logrus.FatalLevel
		default: notifyLevel = logrus.InfoLevel
	}
	Logger.SetFormatter(&logrus.JSONFormatter{})
    Logger.SetLevel(notifyLevel)
	Logger.Printf("Logger set minimum severity is '%s'", notifyLevel.String())
}
