package webserver

import (
	"github.com/nir0k/HighFrequencyDNSChecker/components/log"
)


func Webserver() {

	port, err := GetEnvVariable("WATCHER_WEB_PORT")
	log.AppLog.Info("Failed to read WATCHER_WEB_PORT:")
    if err != nil {
		log.AppLog.Error("Failed to read WATCHER_WEB_PORT:", err)
    }
	
	currentPort = port
	
	done := make(chan bool)
    go StartServer(port, done)
    go WatchForPortChanges(&done)
	<-done
}