package webserver

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
	"fmt"
)

var (
	Timeout int
)

func Webserver() {

	var (
		conf sqldb.WebServerConfiguration
		err error
	)
	conf, err = sqldb.GetWebServerConfig(sqldb.AppDB)
	if err != nil {
		log.AppLog.Error("Failed to get web-server configuration from db:", err)
		fmt.Println("Failed to get web-server configuration from db:", err)
    }

	done := make(chan bool)
    go StartServer(conf, done)
	<-done
}