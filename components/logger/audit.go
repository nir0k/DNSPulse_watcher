package logger

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Audit *logrus.Logger


func AuditSetup(path string, maxSize int, maxFiles int, maxAge int, severity string) {
	Audit = logrus.New()
	Audit.SetOutput(&lumberjack.Logger{
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
	Audit.SetFormatter(&logrus.JSONFormatter{})
    Audit.SetLevel(notifyLevel)
	Audit.Printf("Audit set minimum severity is '%s'", notifyLevel.String())
}
