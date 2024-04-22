package datastore

import (
	"DNSPulse_watcher/pkg/logger"
	"DNSPulse_watcher/pkg/tools"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type LogAppConfigStruct struct {
	Path 		string	
	MinSeverity string	
	MaxAge 		int		
	MaxSize 	int		
	MaxFiles 	int		
}

type LocalConfStruct struct {
	Location 			string	`yaml:"location"`
	SecurityZone 		string	`yaml:"securityZone"`
    Hostname            string
    IPAddress           string
}

type ConfigHubConfigStruct struct {
    Host        string  `yaml:"host"`
    Port        int     `yaml:"port"`
    Segment     string  `yaml:"segment"`
    Token       string  `yaml:"token"`
    EncryptKey  string  `yaml:"encryptionKey"`
    Path        string  `yaml:"path"`
}

type LocalConfigurationStruct struct {
	LocalConf   LocalConfStruct         `yaml:"General"`
	AppLogger 	LogAppConfigStruct
	ConfigHUB   ConfigHubConfigStruct   `yaml:"ConfigHub"`
}

var (
	localConfig LocalConfigurationStruct
	localConfigMutex sync.RWMutex
	localConfigFile string
    localConfLastHash string
)

func SetLocalConfigFilePath(path string){
	localConfigMutex.Lock()
    localConfigFile = path
    localConfigMutex.Unlock()
}

func SetLogConfig(logconfig LogAppConfigStruct){
	localConfigMutex.Lock()
    localConfig.AppLogger = logconfig
    localConfigMutex.Unlock()
}

func GetConfHash() string {
	localConfigMutex.RLock()
    defer localConfigMutex.RUnlock()
    return localConfLastHash
}

// func SetNewPollingHash(hash string) {
//     localConfigMutex.Lock()
//     defer localConfigMutex.Unlock()
//     localConfLastHash
// }

func LoadLocalConfig() (bool, error) {
	fileData, err := os.ReadFile(localConfigFile)
    if err != nil {
        return false, err
    }
    newHash, err := tools.CalculateHash(string(localConfigFile))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", localConfigFile, err)
        return false, err
    }
    if localConfLastHash == newHash {
        logger.Logger.Debug("Configuration file has not been changed")
        return false, nil
    }
    logger.Logger.Infof("Configuration file has been changed")

	var newLocalConfig LocalConfigurationStruct
    if err := yaml.Unmarshal(fileData, &newLocalConfig); err != nil {
        return false, err
    }
    newLocalConfig.LocalConf.Hostname, err = os.Hostname()
    if err != nil {
		newLocalConfig.LocalConf.Hostname = "Current host"
	}
    newLocalConfig.LocalConf.IPAddress = tools.GetLocalIP()

    newLocalConfig.AppLogger = localConfig.AppLogger

	localConfigMutex.Lock()
    localConfig = newLocalConfig
    localConfigMutex.Unlock()
	localConfLastHash = newHash

    return true, nil
}

func GetLocalConfig() *LocalConfigurationStruct {
    localConfigMutex.RLock()
    defer localConfigMutex.RUnlock()
    return &localConfig
}