package config

import (
	"HighFrequencyDNSChecker/components/db"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)


func Setup() error {
	var (
		err error
		isChange bool
		resolvers []sqldb.Resolver
	)
	filepath := "config.yaml"

	mainConf, err := LoadMainConfig(filepath)
	if err != nil {
		fmt.Printf("Error loading configuration from '%s' file, error: %v\n", filepath, err)
		return err
	}
    sqldb.DBName = mainConf.General.DBname
	sqldb.AppDB, err = sql.Open("sqlite3", mainConf.General.DBname)
    if err != nil {
        fmt.Printf("Error opening database: %v\n", err)
        return err
    }
    // defer sqldb.AppDB.Close()

	err = sqldb.InitDB(sqldb.AppDB)
	if err != nil {
		fmt.Println("Error init DB, error: ", err)
		return err
	}


	hostname, err := os.Hostname()
	if err != nil {
		// log.Fatalf("Error getting hostname: %v", err)
		hostname = "Current host"
	}
	ip_address := GetLocalIP()

	isChange, err = IsMainConfigChange(sqldb.AppDB)
	if err != nil {
		fmt.Println("Error IsMainConfigChange, error: ", err)
		return err
	}
	current_time := time.Now().Unix()
	if isChange {
		mainConf.General.IPAddress = ip_address
		mainConf.General.Hostname = hostname
		err = sqldb.InsertMainConfig(sqldb.AppDB, mainConf.General)
		if err != nil {
			fmt.Println("Error InsertMainConfig, error: ", err)
			return err
		}
		err = sqldb.InsertLogConfig(sqldb.AppDB, mainConf.Log, 0)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertLogConfig(sqldb.AppDB, mainConf.Audit, 1)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		
		err = sqldb.InsertWebserverConfig(sqldb.AppDB, mainConf.WebServer)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertSyncConfig(sqldb.AppDB, mainConf.Sync)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertPrometheusConfig(sqldb.AppDB, mainConf.Prometheus)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertPrometeusLabelsConfig(sqldb.AppDB, mainConf.PrometheusLabels)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		mainConf.Resolvers, err = SetAdditionalInfoForResolvers(mainConf.Resolvers, current_time)
		if err != nil {
			fmt.Println("Error GetAdditionalInfoforResolvers, error: ", err)
			return err
		}
		err = sqldb.InsertResolversConfig(sqldb.AppDB, mainConf.Resolvers)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}

		resolvers, err = ReadResolversFromCSV(mainConf.Resolvers)
		if err != nil {
			fmt.Println("Error ReadResolversFromCSV, error: ", err)
			return err
		}

		err = sqldb.InsertResolvers(sqldb.AppDB, resolvers)
		if err != nil {
			fmt.Println("Error InsertResolvers, error: ", err)
			return err
		}

		own := sqldb.MemberConfiguration{
			SyncID: 1,
			Hostname: hostname,
			Port: mainConf.WebServer.Port,
			IPAddress: ip_address,
			Location: mainConf.Watcher.Location,
			SecurityZone: mainConf.Watcher.SecurityZone,
			SeverLastCheck: current_time,
			ConfigHash: mainConf.General.Hash,
			ConfigLastCheck: current_time,
			ConfigLastUpdate: current_time,
			ResolvHash: mainConf.Resolvers.Hash,
			ResolvLastCheck: current_time,
			ResolvLastUpdate: current_time,
			IsLocal: true,
	
		}
		sqldb.UpsertSyncMember(sqldb.AppDB, own)
		log.Println("Configuration inserted into database")
	} else {
		log.Println("Configuration don't changed")
		sqldb.UpdateLastCheck(sqldb.AppDB, sqldb.DBName, current_time)
		sqldb.UpdateSyncMemberTimestamps(sqldb.AppDB, hostname, current_time, 0, current_time, 0, current_time, mainConf.Sync.IsEnable)
	}

	return nil
}

func IsMainConfigChange(db *sql.DB) (bool, error) {
	
	config, err := sqldb.GetMainConfig(db)
    if err != nil {
        // log.Fatal(err)
		return false, err
    }
	if config == (sqldb.MainConfiguration{}) {
		return true, nil
	}
	currentHash, err := CalculateHash(config.ConfPath, md5.New)
	if err != nil {
		return false, err
    }
	// fmt.Printf("Main Config: %+v\n", config)
	return currentHash != config.Hash, nil
}

func IsResolversListChange(db *sql.DB) (bool, error) {
	config, err := sqldb.GetResolverConfig(db)
	if err != nil {
        // log.Fatal(err)
		return false, err
    }
	if config == (sqldb.ResolversConfiguration{}) {
		return true, nil
	}
	currentHash, err := CalculateHash(config.Path, md5.New)
	if err != nil {
		return false, err
    }
	// fmt.Printf("Main Config: %+v\n", config)
	return currentHash != config.Hash, nil
}

func CheckConfig(db *sql.DB) (bool, error) {
	var (
		err error
		conf sqldb.Config
		// isChange bool
		currentHash string
	)
	
	conf, err = sqldb.GetConfgurations(db)
	currentHash, err = CalculateHash(conf.General.ConfPath, md5.New)
	if err != nil {
		return false, err
    }
	if currentHash == conf.General.Hash {
		err = UpdateLastCheckTimestamps(db, conf)
		return false, err
	}

	err = UpdateConfiguration(db, conf)
	return true, err
}

func UpdateConfiguration(db *sql.DB, mainConf sqldb.Config) error {
	var (
		err error
	)
	current_time := time.Now().Unix()
	err = sqldb.InsertMainConfig(db, mainConf.General)
		if err != nil {
			fmt.Println("Error InsertMainConfig, error: ", err)
			return err
		}
		err = sqldb.InsertLogConfig(db, mainConf.Log, 0)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertLogConfig(db, mainConf.Audit, 1)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		
		err = sqldb.InsertWebserverConfig(db, mainConf.WebServer)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertSyncConfig(db, mainConf.Sync)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertPrometheusConfig(db, mainConf.Prometheus)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		err = sqldb.InsertPrometeusLabelsConfig(db, mainConf.PrometheusLabels)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}
		mainConf.Resolvers, err = SetAdditionalInfoForResolvers(mainConf.Resolvers, current_time)
		if err != nil {
			fmt.Println("Error GetAdditionalInfoforResolvers, error: ", err)
			return err
		}
		err = sqldb.InsertResolversConfig(db, mainConf.Resolvers)
		if err != nil {
			fmt.Println("Error InsertLogConfig, error: ", err)
			return err
		}

		err = sqldb.UpdateMainConfigTimestamps(db, mainConf.General.DBname, current_time, current_time)
		if err != nil {
			fmt.Println("Error UpdateMainConfigTimestamps, error: ", err)
			return err
		}

		err = sqldb.UpdateSyncMemberTimestamps(db, mainConf.General.Hostname, current_time, current_time, current_time, current_time, current_time, mainConf.Sync.IsEnable)
		if err != nil {
			fmt.Println("Error UpdateSyncMemberTimestamps, error: ", err)
			return err
		}

		err = sqldb.UpdateResolversConfigTimestamps(db, mainConf.Resolvers.Path, current_time, current_time)
		if err != nil {
			fmt.Println("Error UpdateSyncMemberTimestamps, error: ", err)
			return err
		}
		return nil
}

func UpdateLastCheckTimestamps(db *sql.DB, mainConf sqldb.Config) error {
	var err error
	current_time := time.Now().Unix()
	err = sqldb.UpdateMainConfigTimestamps(db, mainConf.General.DBname, current_time, 0)
		if err != nil {
			fmt.Println("Error UpdateMainConfigTimestamps, error: ", err)
			return err
		}

		err = sqldb.UpdateSyncMemberTimestamps(db, mainConf.General.Hostname, current_time, 0, current_time, 0, current_time, mainConf.Sync.IsEnable)
		if err != nil {
			fmt.Println("Error UpdateSyncMemberTimestamps, error: ", err)
			return err
		}

		err = sqldb.UpdateResolversConfigTimestamps(db, mainConf.Resolvers.Path, current_time, 0)
		if err != nil {
			fmt.Println("Error UpdateSyncMemberTimestamps, error: ", err)
			return err
		}
		return nil
}