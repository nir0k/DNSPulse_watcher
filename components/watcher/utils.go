package watcher

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
)

func GetPrometheusConfigsFromDB () error {
	var (
		err error
	)
	WatcherConfig, err = sqldb.GetWatcherConfig(sqldb.AppDB)
    if err != nil {
		log.AppLog.Error("Failed get watcher config from db:", err)
		return err
	}
    MainConfig, err = sqldb.GetMainConfig(sqldb.AppDB)
    if err != nil {
		log.AppLog.Error("Failed get main config from db:", err)
		return err
	}
	PrometheusConfig, err = sqldb.GetPrometheusConfig(sqldb.AppDB)
    if err != nil {
		log.AppLog.Error("Failed get prometheus config from db:", err)
		return err
	}
	PrometheusLabel, err = sqldb.GetPrometheusLabelConfig(sqldb.AppDB)
    if err != nil {
		log.AppLog.Error("Failed get prometheus labels config from db:", err)
		return err
	}

	return nil
}