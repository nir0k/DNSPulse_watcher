package main

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	syncing "HighFrequencyDNSChecker/components/sync"
	"HighFrequencyDNSChecker/components/tools"

	"HighFrequencyDNSChecker/components/watcher"
	"HighFrequencyDNSChecker/components/webserver"
	"flag"
	"os"
	"time"
)

func setup() {
	var configFilePath = flag.String("config", "config.yaml", "Path to the configuration file")
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

	if !tools.FileExists(*configFilePath) {
		logger.Logger.Fatalf("Configuration file '%s' not exist", *configFilePath)
	}
	datastore.SetConfigFilePath(*configFilePath)
	logConf := datastore.LogAppConfigStruct {
		Path: *logPath,
		MinSeverity: *logSeverity,
		MaxAge: *logMaxAge,
		MaxSize: *logMaxSize,
		MaxFiles: *logMaxFiles,
	}
	datastore.SetLogConfig(logConf)
	logger.Logger.Infof("HighFrequencyDNSChecker started with configuration from '%s'\n", *configFilePath)
}

func main() {
	setup()
	_, err := datastore.LoadConfig()
	auditConf := datastore.GetConfig().AuditLogger
	logger.AuditSetup(auditConf.Path, auditConf.MaxSize, auditConf.MaxFiles, auditConf.MaxAge, auditConf.MinSeverity)
	if err != nil {
		logger.Logger.Fatalf("Error to load configuration, please check config file. error: %s\n", err)
	}
	_, err = datastore.ReadResolversFromCSV()
	if err != nil {
		logger.Logger.Fatalf("Error reading or updating CSV file: %s", err)
	}

	go webserver.Webserver()


	ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for range ticker.C {
            confState, err := datastore.LoadConfig()
            if err != nil {
                logger.Logger.Errorf("Error reloading configuration: %s\n", err)
            } else {
                logger.Logger.Debug("Configuration reloaded successfully")
            }
			csvState, err := datastore.ReadResolversFromCSV()
			if err != nil {
				logger.Logger.Errorf("Error reading or updating CSV file: %s", err)
			} else {
				logger.Logger.Debug("PollingHosts updated successfully")
			}
			if csvState || confState {
				polling.CreatePolling()
			}
			syncing.CheckMembersConfig()
        }
    }()
	
	polling.CreatePolling()
	
	select {}
}

