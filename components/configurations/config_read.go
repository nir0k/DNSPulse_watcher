package config

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"crypto/md5"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)


func LoadMainConfig(filepath string) (sqldb.Config, error){
	var config sqldb.Config

    // Read YAML file
    data, err := os.ReadFile(filepath)
    if err != nil {
        return sqldb.Config{}, err
    }

    // Parse YAML
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        // log.Fatalf("error: %v", err)
		return sqldb.Config{}, err
    }

    config.General.ConfPath = filepath
    config.General.Hash, err = CalculateHash(filepath, md5.New)
    config.General.LastCheck = time.Now().Unix()
    config.General.LastUpdate = config.General.LastCheck
    if err != nil {
        return sqldb.Config{}, err
    }
    // Now you can use `config` struct
    // log.Printf("Config: %+v", config)
	return config, nil
}

func SetAdditionalInfoForResolvers(conf sqldb.ResolversConfiguration, timestamp int64) (sqldb.ResolversConfiguration, error) {
    var (
        err error
        hash string
    )
    conf.LastCheck = timestamp
    hash, err = CalculateHash(conf.Path, md5.New)
    if err != nil {
        return conf, err
    }
    conf.LastUpdate = timestamp
    conf.Hash = hash

    return conf, nil
}
