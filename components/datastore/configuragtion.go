package datastore

import (
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

//  Store configuration after start app
type LogAppConfigStruct struct {
	Path 		string	
	MinSeverity string	
	MaxAge 		int		
	MaxSize 	int		
	MaxFiles 	int		
}

type LogAuditConfigStruct struct {
	Path 		string	`yaml:"path"`
	MinSeverity string	`yaml:"minSeverity"`
	MaxAge 		int		`yaml:"maxAge"`
	MaxSize 	int		`yaml:"maxSize"`
	MaxFiles 	int		`yaml:"maxFiles"`
}

type GeneralConfigStruct struct {
	ConfigCheckInterval int		`yaml:"confCheckInterval"`
	Location 			string	`yaml:"location"`
	SecurityZone 		string	`yaml:"securityZone"`
    Hostname            string
    IPAddress           string
}

type WebServerConfigStruct struct {
	Port 			int		`yaml:"port"`
	ListenAddress 	string	`yaml:"listenAddress"`
	SSLEnabled 		bool	`yaml:"sslIsEnable"`
	SSLCertPath 	string	`yaml:"sslCertPath"`
	SSLKeyPath 		string	`yaml:"sslKeyPath"`
	SesionTimeout 	int		`yaml:"sesionTimeout"`
	Username 		string	`yaml:"username"`
	Password 		string	`yaml:"password"`
}

type SyncConfigStruct struct {
	SyncEnabled bool	`yaml:"isEnable"`
	Token 		string	`yaml:"token"`
}

type SyncMembersStruct struct {
	Host string	`yaml:"hostname"`
	Port int	`yaml:"port"`
}

type PrometheusConfStruct struct {
	URL 			string						`yaml:"url"`
	AuthEnabled 	bool						`yaml:"auth"`
	Username 		string						`yaml:"username"`
	Password 		string						`yaml:"password"`
	MetricName 		string						`yaml:"metricName"`
	RetriesCount 	int							`yaml:"retriesCount"`
	BufferSize 		int							`yaml:"buferSize"`
	Labels 			PrometheusLabelConfigStruct `yaml:"labels"`
}

type PrometheusLabelConfigStruct struct {
	Opcode             bool `yaml:"opcode"`
    Authoritative      bool `yaml:"authoritative"`
    Truncated          bool `yaml:"truncated"`
    Rcode              bool `yaml:"rcode"`
    RecursionDesired   bool `yaml:"recursionDesired"`
    RecursionAvailable bool `yaml:"recursionAvailable"`
    AuthenticatedData  bool `yaml:"authenticatedData"`
    CheckingDisabled   bool `yaml:"checkingDisabled"`
    PollingRate        bool `yaml:"pollingRate"`
    Recursion          bool `yaml:"recursion"`
}

type PollingConfigStruct struct {
	Path 			string	`yaml:"path"`
	Delimeter 		string	`yaml:"delimeter"`
	ExtraDelimeter 	string	`yaml:"extraDelimeter"`
	PollTimeout 	int		`yaml:"pullTimeout"`
}

type ConfigurationStruct struct {
	General 	GeneralConfigStruct 	`yaml:"General"`
	AppLogger 	LogAppConfigStruct		
	AuditLogger LogAuditConfigStruct	`yaml:"Audit"`
	WebServer 	WebServerConfigStruct	`yaml:"WebServer"`
	Sync 		SyncConfigStruct		`yaml:"Sync"`
	SyncMembers []SyncMembersStruct		`yaml:"SyncMembers"`
	Prometheus 	PrometheusConfStruct	`yaml:"Prometheus"`
	Polling 	PollingConfigStruct		`yaml:"Resolvers"`
}

type HashStruct struct {
    LastHash    string  `json:"Hash"`
    LastUpdate  int64   `json:"LastUpdate"`
}

type StateStruct struct {
	Configuration 	HashStruct	`json:"Configuration"`
	Csv				HashStruct	`json:"Csv"`
}

type SyncMemberStateStruct struct {
	State StateStruct
	LastCheckDate int64
    Hostname string
    Port int
    IsLocal bool
}

var (
	config ConfigurationStruct
	configMutex sync.RWMutex
	configFile string
    lastHash HashStruct
    syncMembersState []SyncMemberStateStruct
)

func SetSyncMembersState(members []SyncMemberStateStruct) {
    configMutex.Lock()
    syncMembersState = members
    configMutex.Unlock()
}

func GetSyncMembersState() []SyncMemberStateStruct {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return syncMembersState
}

func SetConfigFilePath(path string){
	configMutex.Lock()
    configFile = path
    configMutex.Unlock()
}

func GetConfigFilePath() string{
	configMutex.RLock()
    defer configMutex.RUnlock()
    return configFile
}

func GetConfHash() HashStruct{
	configMutex.RLock()
    defer configMutex.RUnlock()
    return lastHash
}

func SetLogConfig(logconfig LogAppConfigStruct){
	configMutex.Lock()
    config.AppLogger = logconfig
    configMutex.Unlock()
}


func LoadConfig() (bool, error) {
	fileData, err := os.ReadFile(configFile)
    if err != nil {
        return false, err
    }
    newHash, err := tools.CalculateHash(string(configFile))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", configFile, err)
        return false, err
    }
    if lastHash.LastHash == newHash {
        logger.Logger.Debug("Configuration file has not been changed")
        return false, nil
    }
    logger.Logger.Infof("Configuration file has been changed")

	var newConfig ConfigurationStruct
    if err := yaml.Unmarshal(fileData, &newConfig); err != nil {
        return false, err
    }
    newConfig.General.Hostname, err = os.Hostname()
    if err != nil {
		newConfig.General.Hostname = "Current host"
	}
    newConfig.General.IPAddress = tools.GetLocalIP()

    newConfig.AppLogger = config.AppLogger
	configMutex.Lock()
    config = newConfig
    configMutex.Unlock()
	lastHash.LastHash = newHash
    lastHash.LastUpdate = time.Now().Unix()

    return true, nil
}

func GetConfig() *ConfigurationStruct {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return &config
}

func GetConfigCopy() ConfigurationStruct {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return config
}

func UpdateGeneralConfig(newGeneralConfig GeneralConfigStruct) error {
    configMutex.Lock()
    config.General = newGeneralConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdateAppLoggerConfig(newConfig LogAppConfigStruct) error {
    configMutex.Lock()
    config.AppLogger = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdateAuditLoggerConfig(newConfig LogAuditConfigStruct) error {
    configMutex.Lock()
    config.AuditLogger = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdateWebServerConfig(newConfig WebServerConfigStruct) error {
    configMutex.Lock()
    config.WebServer = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdateSyncConfig(newConfig SyncConfigStruct) error {
    configMutex.Lock()
    config.Sync = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdatePrometheusConfig(newConfig PrometheusConfStruct) error {
    configMutex.Lock()
    config.Prometheus = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdatePollingConfig(newConfig PollingConfigStruct) error {
    configMutex.Lock()
    config.Polling = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func UpdateSyncMembersConfig(newConfig []SyncMembersStruct) error {
    configMutex.Lock()
    config.SyncMembers = newConfig
    configMutex.Unlock()

    return SaveConfigToFile()
}

func SaveConfigToFile() error {
    configMutex.RLock()
    defer configMutex.RUnlock()

    fileData, err := yaml.Marshal(config)
    if err != nil {
        return err
    }

    // Write to a temporary file first
    tempFile := configFile + ".tmp"
    if err := os.WriteFile(tempFile, fileData, 0644); err != nil {
        return err
    }

    // Rename temporary file to the actual config file
    return os.Rename(tempFile, configFile)
}