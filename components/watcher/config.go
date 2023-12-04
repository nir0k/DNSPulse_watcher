package watcher

import (
	"crypto/md5"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	// log "github.com/sirupsen/logrus"
	"github.com/nir0k/HighFrequencyDNSChecker/components/log"
)


var (
    Dns_param dns_param
    Log_conf log_conf   
    Config config
    DnsServers []Resolver
)


func Setup() {
    Config.Conf_path = ".env"
    state := ReadConfig(Config.Conf_path)
    if !state {
        fmt.Println("Error load configuration parametrs. check config in .env files")
        log.AppLog.Fatal("Error load configuration parametrs. check config in .env files")
    }
    state = readDNSServersFromCSV()
    if !state {
        fmt.Println("Error load dns server list from file '", Dns_param.Dns_servers_path, "'. check config in .env files")
        log.AppLog.Fatal("Error load dns server list from file '", Dns_param.Dns_servers_path, "'. check config in .env files")
    } else {
        log.AppLog.Info("DNS info: DNS server count:", len(DnsServers) , ", answer_timeout:", Dns_param.Timeout)
    }
}


func CheckConfig() (bool, bool) {
    conf_compare, _ := compareFileHash(Config.Conf_path, Config.Conf_md5hash)
    if !conf_compare {
        sl :=  log.AppLog.GetLevel()
        log.AppLog.SetLevel(logrus.InfoLevel)
        log.AppLog.Info("Config has been changed")
        fmt.Println("Config has been changed")
        log.AppLog.SetLevel(sl)
        state := ReadConfig(Config.Conf_path)
        if !state {
            log.AppLog.Warn("New config in '", Config.Conf_path, "' is wrong. Use old config")
        } else {
            log.AppLog.Info("New config was loaded")
        }
    } else {
        log.AppLog.Debug("No new config")
    }

    resolvers_state, _ := compareFileHash(Dns_param.Dns_servers_path, Dns_param.Dns_servers_file_md5hash)
    if !resolvers_state {
        sl := log.AppLog.GetLevel()
        log.AppLog.SetLevel(logrus.InfoLevel)
        log.AppLog.Info("List of DNS service has been changed")
        log.AppLog.SetLevel(sl)
        state := readDNSServersFromCSV()
        if !state {
            log.AppLog.Warn("New List of DNS servers is wrong. Use old list of DNS service")
        } else {
            log.AppLog.SetLevel(logrus.InfoLevel)
            log.AppLog.Info("New list of DNS servers was loaded")
            log.AppLog.Info("DNS info: DNS server count:", len(DnsServers) , ", answer_timeout:", Dns_param.Timeout)
            log.AppLog.SetLevel(sl)
            CreatePolling()
        }
    }
    return conf_compare, resolvers_state
}


func ReadConfig(filePath string) bool {
    
    err := godotenv.Load(filePath)
    if err != nil {
        fmt.Println("Error loading ", Config.Conf_path, " file", err)
        log.AppLog.Error("Error loading ", Config.Conf_path, " file", err)
        return false
    }

    if !readConfigLog() {
        return false
    }

    if !readConfigWatcher(filePath) {
        return false
    }

    if !readConfigPrometheus(filePath) {
        return false
    }

    if !readConfigDNS(filePath) {
        return false
    }

    return true
}


func readConfigLog() bool {
    var new_log_conf log_conf

    new_log_conf.Log_path = os.Getenv("LOG_FILE")
    validPathRegex := regexp.MustCompile("^[a-zA-Z0-9-_/.]+$")
    if !validPathRegex.MatchString(new_log_conf.Log_path) {
        fmt.Println("Error create/open log file ", new_log_conf.Log_path)
        log.AppLog.Error("Error create/open log file '", new_log_conf.Log_path, "'.")
        return false
    }

    new_log_conf.Log_level = os.Getenv("LOG_LEVEL")
    notifyLevel := logrus.InfoLevel
    switch new_log_conf.Log_level {
        case "debug": notifyLevel = logrus.DebugLevel
        case "info": notifyLevel = logrus.InfoLevel
        case "warning": notifyLevel = logrus.WarnLevel
        case "error": notifyLevel = logrus.ErrorLevel
        case "fatal": notifyLevel = logrus.FatalLevel
        default: {
            logrus.Error("Error min log severity '", new_log_conf.Log_level, "'.")
            return false
        } 
    }
    log.InitAppLogger(new_log_conf.Log_path, notifyLevel)

    Log_conf = new_log_conf
    return true
}


func readConfigWatcher(filePath string) bool {
    var (
        new_config config
        err error
    )

    new_config.Conf_path = ".env"

    new_config.Ip, err = getLocalIP()
	if err != nil {
		log.AppLog.Error("Error getting watcher IP address:", err)
        return false
	}
    new_config.Buffer_size, err = strconv.Atoi(os.Getenv("BUFFER_SIZE"))
    if err != nil {
        log.AppLog.Error("Warning: Variable BUFFER_SIZE is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    new_config.Conf_md5hash, err = calculateHash(filePath, md5.New)
    if err != nil {
        log.AppLog.Error("Error: calculate hash to file '", filePath, "'")
        return false
    }

    new_config.Check_interval, err = strconv.Atoi(os.Getenv("CONF_CHECK_INTERVAL"))
    if err != nil {
        log.AppLog.Error("Warning: Variable CONF_CHECK_INTERVAL is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    new_config.Location = os.Getenv("WATCHER_LOCATION")
    if len(new_config.Location) == 0 {
        log.AppLog.Error("Error: Variable WATCHER_LOCATION is required in ", filePath, " file. Please add this variable with value")
        return false
    }

    new_config.SecurityZone = os.Getenv("WATCHER_SECURITYZONE")
    if len(new_config.SecurityZone) == 0 {
        log.AppLog.Error("Error: Variable WATCHER_SECURITYZONE is required in ", filePath, " file. Please add this variable with value")
        return false
    }

    new_config.Hostname, err = os.Hostname()
    if err != nil {
        log.AppLog.Error("Error getting watcher hostname:", err)
        return false
    }

    Config = new_config
    return true
}


func readConfigPrometheus(filePath string) bool {
    var (
        new_prometheus prometheus
        err error
    )

    new_prometheus.Url = os.Getenv("PROM_URL")
    if !isValidURL(new_prometheus.Url) {
        log.AppLog.Error("Error: Variable PROM_URL is required in ", filePath, " file. Please add this variable with value")
        return false
    }

    new_prometheus.Metric = os.Getenv("PROM_METRIC")
    if !isAlphaNumericWithDashOrUnderscore(new_prometheus.Metric) {
        log.AppLog.Error("Error: Variable PROM_METRIC is empty or wrong in ", filePath, " file.")
        return false
    }

    new_prometheus.Retries, err = strconv.Atoi(os.Getenv("PROM_RETRIES"))
    if err != nil {
        log.AppLog.Error("Error: Variable PROM_RETRIES is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    new_prometheus.Auth, err = strconv.ParseBool(os.Getenv("PROM_AUTH"))
    if err != nil {
        log.AppLog.Error("Error: Variable PROM_AUTH is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    if new_prometheus.Auth {
        new_prometheus.Username = os.Getenv("PROM_USER")
        if len(new_prometheus.Username) == 0 {
            log.AppLog.Error("Error: Variable PROM_USER is required in ", filePath, " file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
        new_prometheus.Password = os.Getenv("PROM_PASS")
        if len(new_prometheus.Password) == 0 {
            log.AppLog.Error("Error: Variable PROM_PASS is required in ", filePath, " file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
    } 

    new_prometheus.Metrics.Opscodes, err = strconv.ParseBool(os.Getenv("OPCODES"))
    if err != nil {
        log.AppLog.Error("Error: Variable OPCODES is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.Authoritative, err = strconv.ParseBool(os.Getenv("AUTHORITATIVE"))
    if err != nil {
        log.AppLog.Error("Error: Variable AUTHORITATIVE is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.Truncated, err = strconv.ParseBool(os.Getenv("TRUNCATED"))
    if err != nil {
        log.AppLog.Error("Error: Variable TRUNCATED is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.Rcode, err = strconv.ParseBool(os.Getenv("RCODE"))
    if err != nil {
        log.AppLog.Error("Error: Variable RCODE is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.RecursionDesired, err = strconv.ParseBool(os.Getenv("RECURSION_DESIRED"))
    if err != nil {
        log.AppLog.Error("Error: Variable RECURSION_DESIRED is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.RecursionAvailable, err = strconv.ParseBool(os.Getenv("RECURSION_AVAILABLE"))
    if err != nil {
        log.AppLog.Error("Error: Variable RECURSION_AVAILABLE is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.AuthenticatedData, err = strconv.ParseBool(os.Getenv("AUTHENTICATE_DATA"))
    if err != nil {
        log.AppLog.Error("Error: Variable AUTHENTICATE_DATA is empty or wrong in ", filePath, " file. error:", err)
        return false
    }
    new_prometheus.Metrics.CheckingDisabled, err = strconv.ParseBool(os.Getenv("CHECKING_DISABLED"))
    if err != nil {
        log.AppLog.Error("Error: Variable CHECKING_DISABLED is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    new_prometheus.Metrics.Polling_rate, err = strconv.ParseBool(os.Getenv("POLLING_RATE"))
    if err != nil {
        log.AppLog.Error("Error: Variable POLLING_RATE is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    new_prometheus.Metrics.Recusrion, err = strconv.ParseBool(os.Getenv("RECURSION"))
    if err != nil {
        log.AppLog.Error("Error: Variable RECURSION is empty or wrong in ", filePath, " file. error:", err)
        return false
    }

    Prometheus = new_prometheus
    return true
}


func readConfigDNS(filePath string) bool {
    var (
        new_dns_param dns_param
        err error
    )

    new_dns_param.Dns_servers_path = os.Getenv("DNS_RESOLVERPATH")
    validRPathRegex := regexp.MustCompile("^[a-zA-Z0-9-_/.]+$")
    if !validRPathRegex.MatchString(new_dns_param.Dns_servers_path) {
        fmt.Println("Error: Variable DNS_RESOLVERPATH is wrong check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")
        log.AppLog.Error("Error: Variable DNS_RESOLVERPATH is wrong check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")
        return false
    }

    new_dns_param.Timeout, err = strconv.Atoi(os.Getenv("DNS_TIMEOUT"))
    if err != nil {
        log.AppLog.Error("Error: Variable DNS_TIMEOUT is empty or wrong in ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.error:", err)
        return false
    }

    delimeter := os.Getenv("DELIMETER")
    if len(delimeter) == 1 {
        new_dns_param.Delimeter = rune(delimeter[0])
    } else {
        log.AppLog.Error("Error: Variable DELIMETER is not a single character. Check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")
        fmt.Println("Error: Variable DELIMETER is not a single character. Check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")

        return false
    }

    new_dns_param.Delimeter_for_additional_field = os.Getenv("DELIMETER_FOR_ADDITIONAL_PARAM")
    if len(new_dns_param.Delimeter_for_additional_field) < 1 {
        fmt.Println("Error: Variable DELIMETER_FOR_ADDITIONAL_PARAM is wrong check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")
        log.AppLog.Error("Error: Variable DELIMETER_FOR_ADDITIONAL_PARAM is wrong check ", filePath, " file. Path:'", new_dns_param.Dns_servers_path, "'.")
        return false
    }

    new_dns_param.Dns_servers_file_md5hash = Dns_param.Dns_servers_file_md5hash

    Dns_param = new_dns_param
    return true
}


func readDNSServersFromCSV() bool {
    dns_list := []Csv{}
    clientsFile, err := os.OpenFile(Dns_param.Dns_servers_path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.AppLog.Error("Error read file ", Dns_param.Dns_servers_path,"| error: ", err)
        return false
	}
	defer clientsFile.Close()
    gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        r := csv.NewReader(in)
        r.LazyQuotes = true
        r.Comma = Dns_param.Delimeter
        return r
    })
	if err := gocsv.UnmarshalFile(clientsFile, &dns_list); err != nil {
		log.AppLog.Error("Error Unmarshal file ", Dns_param.Dns_servers_path,"| error: ", err)
        return false
	}
    
    DnsServers, _ = parseZones(dns_list)
    
    new_md5hash, err := calculateHash(Dns_param.Dns_servers_path, md5.New)
    if err != nil {
        log.AppLog.Error("Error: calculate hash to file '", Dns_param.Dns_servers_path, "'. error:", err)
        return false
    }
    Dns_param.Dns_servers_file_md5hash = new_md5hash
    return true
}


func parseZones(records []Csv) ([]Resolver, error) {
    var resolvers []Resolver
    for _, record := range records {
        
        zoneNames :=  strings.Split(record.Zonename, Dns_param.Delimeter_for_additional_field)
        queryRPSs := strings.Split(record.Query_count_rps, Dns_param.Delimeter_for_additional_field)
        mm_mode, err := strconv.ParseBool(record.Service_mode)
        if err != nil {
            log.AppLog.Warning("Warning: Error parse server mode value for server: '", record.Server, "', value 'service_mode': ", record.Service_mode, "err:", err)
            mm_mode = false
        }
        for i, zonename := range zoneNames {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSs[i])
            if err != nil {
                log.AppLog.Warning("Warning: Error parse query count rps value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 5
            }
            resolver := Resolver{
                Server: record.Server,
                Server_ip: record.Server_ip,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                Server_security_zone: record.Server_security_zone,
                Prefix: record.Prefix,
                Protocol: record.Protocol,
                Zonename: zonename,
                Recursion: false,
                Query_count_rps: queryRPSInt,
                Service_mode: mm_mode,
            }
            resolvers = append(resolvers, resolver)
        }

        zoneNamesRecursion :=  strings.Split(record.Zonename_with_recursion, "&")
        queryRPSsRecursion := strings.Split(record.Query_count_with_recursion, "&")
        for i, zonename := range zoneNamesRecursion {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSsRecursion[i])
            if err != nil {
                log.AppLog.Warning("Warning: Error parse query count rps value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 2
            }
            resolver := Resolver{
                Server: record.Server,
                Server_ip: record.Server_ip,
                Domain: record.Domain,
                Location: record.Location,
                Site: record.Site,
                Server_security_zone: record.Server_security_zone,
                Prefix: record.Prefix,
                Protocol: record.Protocol,
                Zonename: zonename,
                Recursion: true,
                Query_count_rps: queryRPSInt,
                Service_mode: mm_mode,
            }
            resolvers = append(resolvers, resolver)
        }
    }
    return resolvers, nil
}