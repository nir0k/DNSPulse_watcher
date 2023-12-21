package main

import (
	config "HighFrequencyDNSChecker/components/configurations"
	sqldb "HighFrequencyDNSChecker/components/db"
	"HighFrequencyDNSChecker/components/log"
	"HighFrequencyDNSChecker/components/watcher"
	webserver "HighFrequencyDNSChecker/components/web"
	"HighFrequencyDNSChecker/components/web/api"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)


func main() {
	var (
		err error
		conf sqldb.MainConfiguration
	)

	err = config.Setup()
	if err != nil {
		fmt.Printf("Error setup app configuration, error:%v\n", err)
		return
	}
	_, err = log.InitAppLogger()
	if err != nil {
		fmt.Printf("Error init logging, error:%v\n", err)
		return
	}
	_, err = log.InitAuthLog()
	if err != nil {
		fmt.Printf("Error init audit logging, error:%v\n", err)
		return
	}

	sl := log.AppLog.GetLevel()
	log.AppLog.SetLevel(logrus.InfoLevel)
    log.AppLog.Info("Frequency DNS cheker start.")
	log.AppLog.SetLevel(sl)

	go webserver.Webserver()
	// watcher.CreatePolling()

	currentTime := time.Now()
	var startTime = currentTime.Truncate(time.Second).Add(time.Second)
	var duration = startTime.Sub(currentTime)
	time.Sleep(duration)
	conf, err = sqldb.GetMainConfig(sqldb.AppDB)
	if err != nil {
		conf.UpdateInterval = 1
	}
	ticker := time.NewTicker(time.Duration(conf.UpdateInterval) * time.Minute)
	go func() {
        for range ticker.C {
            config.CheckConfig(sqldb.AppDB)
			api.CheckConfFromMembers()
        }
    }()
	watcher.CreatePolling()
    
    select {}

}