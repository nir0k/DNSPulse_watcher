package main

import (
	"DNSPulse_watcher/pkg/datastore"
	grpcclient "DNSPulse_watcher/pkg/gRPC-client"
	"DNSPulse_watcher/pkg/logger"
	"DNSPulse_watcher/pkg/tools"
	polling "DNSPulse_watcher/pkg/watcher"
	"fmt"

	"flag"
	"os"
	"time"
)

func setup() {
	var conf = flag.String("conf", "mainConf.yaml", "Path to the main configuration file")
	var logPath = flag.String("logPath", "log.json", "Path to the log file")
	var logSeverity = flag.String("logSeverity", "debug", "Min log severity")
	var logMaxSize = flag.Int("logMaxSize", 10, "Max size for log file (Mb)")
	var logMaxFiles = flag.Int("logMaxFiles", 10, "Maximum number of log files")
	var logMaxAge = flag.Int("logMaxAge", 10, "Maximum log file age")
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "--help" {
        flag.PrintDefaults()
        return
    }
	logger.LogSetup(*logPath, *logMaxSize, *logMaxFiles, *logMaxAge, *logSeverity)

	if !tools.FileExists(*conf) {
		logger.Logger.Fatalf("Configuration file '%s' not exist", *conf)
	}
	datastore.SetLocalConfigFilePath(*conf)
	logConf := datastore.LogAppConfigStruct {
		Path: *logPath,
		MinSeverity: *logSeverity,
		MaxAge: *logMaxAge,
		MaxSize: *logMaxSize,
		MaxFiles: *logMaxFiles,
	}
	datastore.SetLogConfig(logConf)
	logger.Logger.Infof("HighFrequencyDNSChecker started with configuration from '%s'\n", *conf)
}

func main() {
	setup()
	_, err := datastore.LoadLocalConfig()
	if err != nil {
		logger.Logger.Errorf("error load local config, err: %v", err)
	}
	conf := datastore.GetLocalConfig()
	_, _, err = grpcclient.FetchConfig(conf.ConfigHUB)
	// logger.AuditSetup(conf.AuditLogger.Path, conf.AuditLogger.MaxSize, conf.AuditLogger.MaxFiles, conf.AuditLogger.MaxAge, conf.AuditLogger.MinSeverity)
	if err != nil {
		logger.Logger.Fatalf("Error to load configuration, please check config file. error: %s\n", err)
	}

	ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for range ticker.C {
            _, err := datastore.LoadLocalConfig()
            if err != nil {
                logger.Logger.Errorf("Error reloading configuration: %s\n", err)
            } else {
                logger.Logger.Debug("Configuration reloaded successfully")
            }
			confState, pollingState, _ := grpcclient.FetchConfig(conf.ConfigHUB)			
			if confState || pollingState {
				polling.CreatePolling()
			}
        }
    }()
	polling.CreatePolling()
	layout := "2006-01-02 15:04:05 -0700 MST"
	fmt.Printf("===========================================\n")
	fmt.Printf("Watcher started.\n  Time: %s\n", time.Now().Format(layout))
	fmt.Printf("-------------------------------------------\n")
	fmt.Printf("  IP address\t: %s\n  Host\t\t: %s\n", conf.LocalConf.IPAddress, conf.LocalConf.Hostname)
	fmt.Printf("  Segment\t: %s\n", conf.ConfigHUB.Segment)
	fmt.Printf("  Location\t: %s\n  Segment\t: %s\n", conf.LocalConf.Location, conf.LocalConf.SecurityZone)
	fmt.Printf("  gRPC server\t: %s:%d\n", conf.ConfigHUB.Host, conf.ConfigHUB.Port)
	fmt.Printf("===========================================\n")
	select {}
}

