package log

import "github.com/sirupsen/logrus"

func GetLogSeverity(severity string) (logrus.Level, bool) {
	notifyLevel := logrus.InfoLevel
	switch severity {
		case "debug": notifyLevel = logrus.DebugLevel
		case "info": notifyLevel = logrus.InfoLevel
		case "warning": notifyLevel = logrus.WarnLevel
		case "error": notifyLevel = logrus.ErrorLevel
		case "fatal": notifyLevel = logrus.FatalLevel
		default: {
			return logrus.WarnLevel, false
		}
	}
	return notifyLevel, true
}