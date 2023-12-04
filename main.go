package main

import (
	"github.com/nir0k/HighFrequencyDNSChecker/components/watcher"
    "github.com/nir0k/HighFrequencyDNSChecker/components/web-server"
	"time"
	log "github.com/sirupsen/logrus"
)

var (
    Polling bool
    Polling_chan chan struct{}
)

func main() {
	watcher.Setup()
    sl := log.GetLevel()
    log.SetLevel(log.InfoLevel)
    log.Info("Frequency DNS cheker start.")
    log.Info("Prometheus info: url:", watcher.Prometheus.Url , ", auth:", watcher.Prometheus.Auth, ", username:", watcher.Prometheus.Username, ", metric_name:", watcher.Prometheus.Metric)
    log.Info("Debug level:", sl.String() )
    log.SetLevel(sl)

    currentTime := time.Now()
	var startTime = currentTime.Truncate(time.Second).Add(time.Second)
	var duration = startTime.Sub(currentTime)
	time.Sleep(duration)

    ticker := time.NewTicker(time.Duration(watcher.Config.Check_interval) * time.Minute)
    go func() {
        for range ticker.C {
            watcher.CheckConfig()
        }
    }()
    go webserver.Webserver()
	watcher.CreatePolling()
    
    select {}
}