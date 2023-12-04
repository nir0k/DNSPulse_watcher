package webserver

import (
	"github.com/sirupsen/logrus"
	"github.com/nir0k/HighFrequencyDNSChecker/components/log"
)


func Webserver() {

	port, err := GetEnvVariable("WATCHER_WEB_PORT")
	log.AppLog.Info("Failed to read WATCHER_WEB_PORT:")
    if err != nil {
		log.AppLog.Error("Failed to read WATCHER_WEB_PORT:", err)
    }
	authlog_filepath, err := GetEnvVariable("WATCHER_WEB_AUTH_LOG_FILE")
    if err != nil {
		log.AppLog.Error("Failed to read WATCHER_WEB_AUTH_LOG_FILE:", err)
    }
	Log_level, err := GetEnvVariable("WATCHER_WEB_AUTH_LOG_LEVEL")
	if err != nil {
		log.AppLog.Error("Failed to read WATCHER_WEB_AUTH_LOG_LEVEL:", err)
    }
    notifyLevel := logrus.InfoLevel
    switch Log_level {
        case "debug": notifyLevel = logrus.DebugLevel
        case "info": notifyLevel = logrus.InfoLevel
        case "warning": notifyLevel = logrus.WarnLevel
        case "error": notifyLevel = logrus.ErrorLevel
        case "fatal": notifyLevel = logrus.FatalLevel
        default: {
            logrus.Error("Error min log severity '", Log_level, "'.")
        } 
    }

	log.InitAuthLog(authlog_filepath, notifyLevel)
	if log.AuthLog == nil {
        log.AppLog.Fatal("Failed to initialize authLog")
    }

	currentPort = port
	
	done := make(chan bool)
    go StartServer(port, done)
    go WatchForPortChanges(&done)
	<-done
}