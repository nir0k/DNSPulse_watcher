package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"hash"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/castai/promwrite"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Csv struct {
	Server			            string `csv:"server"`
	Server_ip		            string `csv:"server_ip"`
	Domain			            string `csv:"domain"`
    Location                    string `csv:"location"`
    Site                        string `csv:"site"`
    Server_security_zone        string `csv:"server_security_zone"`
    Prefix                      string `csv:"prefix"`
    Protocol                    string `csv:"protocol"`
    Zonename		            string `csv:"zonename"`
    Query_count_rps             string `csv:"query_count_rps"`
    Zonename_with_recursion     string `csv:"zonename_with_recursion"`
    Query_count_with_recursion  string `csv:"query_count_with_recursion_rps"`
    Maintenance_mode            string `csv:"maintenance_mode"`

}

type Resolver struct {
    server string
	server_ip string
	domain string
    location string
    site string
    server_security_zone string
    prefix string
    protocol string
    zonename string
    recursion bool
    query_count_rps int
    maintenance_mode bool
}


type prometheus struct {
    url string
    auth  bool
    username string
    password string
    metric string
    retries int
    metrics metrics
}


type metrics struct {
    rcode bool
    opscodes bool
    authoritative bool
    truncated bool
    recursionDesired bool
    recursionAvailable bool
    authenticatedData bool
    checkingDisabled bool
    polling_rate bool
    recusrion bool
}


type dns_param struct {
    timeout  int
    dns_servers_path string
    dns_servers_file_md5hash string
    delimeter rune
    delimeter_for_additional_field string
}


type log_conf struct {
    log_path string
    log_level string
}


type config struct {
    conf_path string
    conf_md5hash string
    check_interval int
    buffer_size int
    ip string
    hostname string
    location string
    securityZone string
}


var (
    Prometheus prometheus
    Dns_param dns_param
    Log_conf log_conf   
    Config config
    DnsServers []Resolver
    Buffer []promwrite.TimeSeries
    Mu sync.Mutex
    Polling bool
    Polling_chan chan struct{}
)


func init() {
    Config.conf_path = ".env"
    state := readConfig()
    if !state {
        fmt.Println("Error load configuration parametrs. check config in .env files")
        log.Fatal("Error load configuration parametrs. check config in .env files")
    }
    state = readDNSServersFromCSV()
    if !state {
        fmt.Println("Error load dns server list from file '", Dns_param.dns_servers_path, "'. check config in .env files")
        log.Fatal("Error load dns server list from file '", Dns_param.dns_servers_path, "'. check config in .env files")
    }
}


func parseZones(records []Csv) ([]Resolver, error) {
    var resolvers []Resolver
    for _, record := range records {
        
        zoneNames :=  strings.Split(record.Zonename, Dns_param.delimeter_for_additional_field)
        queryRPSs := strings.Split(record.Query_count_rps, Dns_param.delimeter_for_additional_field)
        mm_mode, err := strconv.ParseBool(record.Maintenance_mode)
        if err != nil {
            log.Warning("Warning: Error parse maitanence mode value for server: '", record.Server, "', value 'maintenance_mode': ", record.Maintenance_mode, "err:", err)
            mm_mode = false
        }
        for i, zonename := range zoneNames {
            if zonename == "" {
                continue
            }
            queryRPSInt, err := strconv.Atoi(queryRPSs[i])
            if err != nil {
                log.Warning("Warning: Error parse maitanence mode value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 5
            }
            resolver := Resolver{
                server: record.Server,
                server_ip: record.Server_ip,
                domain: record.Domain,
                location: record.Location,
                site: record.Site,
                server_security_zone: record.Server_security_zone,
                prefix: record.Prefix,
                protocol: record.Protocol,
                zonename: zonename,
                recursion: false,
                query_count_rps: queryRPSInt,
                maintenance_mode: mm_mode,
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
                log.Warning("Warning: Error parse maitanence mode value for server: '", record.Server, "', value 'query_count_rps': ", record.Query_count_rps, "err:", err)
                queryRPSInt = 2
            }
            resolver := Resolver{
                server: record.Server,
                server_ip: record.Server_ip,
                domain: record.Domain,
                location: record.Location,
                site: record.Site,
                server_security_zone: record.Server_security_zone,
                prefix: record.Prefix,
                protocol: record.Protocol,
                zonename: zonename,
                recursion: true,
                query_count_rps: queryRPSInt,
                maintenance_mode: mm_mode,
            }
            resolvers = append(resolvers, resolver)
        }
    }
    return resolvers, nil
}


func readDNSServersFromCSV() bool {
    dns_list := []Csv{}
    clientsFile, err := os.OpenFile(Dns_param.dns_servers_path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Error("Error read file ", Dns_param.dns_servers_path,"| error: ", err)
        return false
	}
	defer clientsFile.Close()
    gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        r := csv.NewReader(in)
        r.LazyQuotes = true
        r.Comma = Dns_param.delimeter
        return r
    })
	if err := gocsv.UnmarshalFile(clientsFile, &dns_list); err != nil {
		log.Error("Error Unmarshal file ", Dns_param.dns_servers_path,"| error: ", err)
        return false
	}
    
    DnsServers, _ = parseZones(dns_list)
    
    new_md5hash, err := calculateHash(Dns_param.dns_servers_path, md5.New)
    if err != nil {
        log.Error("Error: calculate hash to file '", Dns_param.dns_servers_path, "'. error:", err)
        return false
    }
    Dns_param.dns_servers_file_md5hash = new_md5hash
    return true
}


func isValidURL(inputURL string) bool {
    _, err := url.ParseRequestURI(inputURL)
    return err == nil
}


func isAlphaNumericWithDashOrUnderscore(input string) bool {
    validRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
    return validRegex.MatchString(input)
}


func calculateHash(filePath string, hashFunc func() hash.Hash) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := hashFunc()
    _, err = io.Copy(hash, file)
    if err != nil {
        return "", err
    }

    hashSum := hash.Sum(nil)
    return fmt.Sprintf("%x", hashSum), nil
}


func compareFileHash(path string, curent_hash string) (bool, error) {
    
    new_hash, err := calculateHash(path, md5.New)
    if err != nil {
        log.Error("Error: calculating MD5 hash to file '", path, "'. error:", err)
        return false, err
    }
    if curent_hash == new_hash {
        return true, nil
    }
    return false, nil
}


func checkConfig() {
    conf_compare, _ := compareFileHash(Config.conf_path, Config.conf_md5hash)
    if !conf_compare {
        sl :=  log.GetLevel()
        log.Info("Config has been changed")
        fmt.Println("Config has been changed")
        log.SetLevel(sl)
        state := readConfig()
        if !state {
            log.Warn("New config in '", Config.conf_path, "' is wrong. Use old config")
        } else {
            log.Info("New config was loaded")
        }
    }

    resolvers_state, _ := compareFileHash(Dns_param.dns_servers_path, Dns_param.dns_servers_file_md5hash)
    if !resolvers_state {
        sl := log.GetLevel()
        log.Info("List of DNS service has been changed")
        log.SetLevel(sl)
        state := readDNSServersFromCSV()
        if !state {
            log.Warn("New List of DNS servers is wrong. Use old list of DNS service")
        } else {
            log.Info("New list of DNS servers was loaded")
            createPolling()
        }
    }
}


func readConfig() bool {

    err := godotenv.Overload()
    if err != nil {
        fmt.Println("Error loading ", Config.conf_path, " file", err)
        log.Error("Error loading ", Config.conf_path, " file", err)
        return false
    }

    if !readConfigLog() {
        return false
    }

    if !readConfigWatcher() {
        return false
    }

    if !readConfigPrometheus() {
        return false
    }

    if !readConfigDNS() {
        return false
    }

    
    return true
}

func readConfigLog() bool {
    var new_log_conf log_conf

    new_log_conf.log_path = os.Getenv("LOG_FILE")
    validPathRegex := regexp.MustCompile("^[a-zA-Z0-9-_/.]+$")
    if !validPathRegex.MatchString(new_log_conf.log_path) {
        fmt.Println("Error create/open log file ", new_log_conf.log_path)
        log.Error("Error create/open log file '", new_log_conf.log_path, "'.")
        return false
    }
    
    file, err := os.OpenFile(new_log_conf.log_path, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
    if err != nil {
        fmt.Println("Error opening file '", Log_conf.log_path,"': %v", err)
        return false
    }
    log.SetFormatter(&log.JSONFormatter{})
    log.SetOutput(file)
    new_log_conf.log_level = os.Getenv("LOG_LEVEL")
    switch new_log_conf.log_level {
        case "debug": log.SetLevel(log.DebugLevel)
        case "info": log.SetLevel(log.InfoLevel)
        case "warning": log.SetLevel(log.WarnLevel)
        case "error": log.SetLevel(log.ErrorLevel)
        case "fatal": log.SetLevel(log.FatalLevel)
        default: {
            log.Error("Error min log severity '", new_log_conf.log_level, "'.")
            return false
        } 
    }

    Log_conf = new_log_conf
    return true
}


func readConfigWatcher() bool {
    var (
        new_config config
        err error
    )

    new_config.conf_path = ".env"

    new_config.ip, err = getLocalIP()
	if err != nil {
		log.Error("Error getting watcher IP address:", err)
        return false
	}
    new_config.buffer_size, err = strconv.Atoi(os.Getenv("BUFFER_SIZE"))
    if err != nil {
        log.Error("Warning: Variable BUFFER_SIZE is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    new_config.conf_md5hash, err = calculateHash(Config.conf_path, md5.New)
    if err != nil {
        log.Error("Error: calculate hash to file '", Config.conf_path, "'")
        return false
    }

    new_config.check_interval, err = strconv.Atoi(os.Getenv("CONF_CHECK_INTERVAL"))
    if err != nil {
        log.Error("Warning: Variable CONF_CHECK_INTERVAL is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    new_config.location = os.Getenv("WATCHER_LOCATION")
    if len(new_config.location) == 0 {
        log.Error("Error: Variable WATCHER_LOCATION is required in ", Config.conf_path, " file. Please add this variable with value")
        return false
    }

    new_config.securityZone = os.Getenv("WATCHER_SECURITYZONE")
    if len(new_config.securityZone) == 0 {
        log.Error("Error: Variable WATCHER_SECURITYZONE is required in ", Config.conf_path, " file. Please add this variable with value")
        return false
    }

    new_config.hostname, err = os.Hostname()
    if err != nil {
        log.Error("Error getting watcher hostname:", err)
        return false
    }

    Config = new_config
    return true
}


func readConfigPrometheus() bool {
    var (
        new_prometheus prometheus
        err error
    )

    new_prometheus.url = os.Getenv("PROM_URL")
    if !isValidURL(new_prometheus.url) {
        log.Error("Error: Variable PROM_URL is required in ", Config.conf_path, " file. Please add this variable with value")
        return false
    }

    new_prometheus.metric = os.Getenv("PROM_METRIC")
    if !isAlphaNumericWithDashOrUnderscore(new_prometheus.metric) {
        log.Error("Error: Variable PROM_METRIC is empty or wrong in ", Config.conf_path, " file.")
        return false
    }

    new_prometheus.retries, err = strconv.Atoi(os.Getenv("PROM_RETRIES"))
    if err != nil {
        log.Error("Error: Variable PROM_RETRIES is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    new_prometheus.auth, err = strconv.ParseBool(os.Getenv("PROM_AUTH"))
    if err != nil {
        log.Error("Error: Variable PROM_AUTH is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    if new_prometheus.auth {
        new_prometheus.username = os.Getenv("PROM_USER")
        if len(new_prometheus.username) == 0 {
            log.Error("Error: Variable PROM_USER is required in ", Config.conf_path, " file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
        new_prometheus.password = os.Getenv("PROM_PASS")
        if len(new_prometheus.password) == 0 {
            log.Error("Error: Variable PROM_PASS is required in ", Config.conf_path, " file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
    } 

    new_prometheus.metrics.opscodes, err = strconv.ParseBool(os.Getenv("OPCODES"))
    if err != nil {
        log.Error("Error: Variable OPCODES is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.authoritative, err = strconv.ParseBool(os.Getenv("AUTHORITATIVE"))
    if err != nil {
        log.Error("Error: Variable AUTHORITATIVE is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.truncated, err = strconv.ParseBool(os.Getenv("TRUNCATED"))
    if err != nil {
        log.Error("Error: Variable TRUNCATED is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.rcode, err = strconv.ParseBool(os.Getenv("RCODE"))
    if err != nil {
        log.Error("Error: Variable RCODE is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.recursionDesired, err = strconv.ParseBool(os.Getenv("RECURSION_DESIRED"))
    if err != nil {
        log.Error("Error: Variable RECURSION_DESIRED is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.recursionAvailable, err = strconv.ParseBool(os.Getenv("RECURSION_AVAILABLE"))
    if err != nil {
        log.Error("Error: Variable RECURSION_AVAILABLE is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.authenticatedData, err = strconv.ParseBool(os.Getenv("AUTHENTICATE_DATA"))
    if err != nil {
        log.Error("Error: Variable AUTHENTICATE_DATA is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }
    new_prometheus.metrics.checkingDisabled, err = strconv.ParseBool(os.Getenv("CHECKING_DISABLED"))
    if err != nil {
        log.Error("Error: Variable CHECKING_DISABLED is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    new_prometheus.metrics.polling_rate, err = strconv.ParseBool(os.Getenv("POLLING_RATE"))
    if err != nil {
        log.Error("Error: Variable POLLING_RATE is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    new_prometheus.metrics.recusrion, err = strconv.ParseBool(os.Getenv("RECURSION"))
    if err != nil {
        log.Error("Error: Variable RECURSION is empty or wrong in ", Config.conf_path, " file. error:", err)
        return false
    }

    Prometheus = new_prometheus
    return true
}


func readConfigDNS() bool {
    var (
        new_dns_param dns_param
        err error
    )

    new_dns_param.dns_servers_path = os.Getenv("DNS_RESOLVERPATH")
    validRPathRegex := regexp.MustCompile("^[a-zA-Z0-9-_/.]+$")
    if !validRPathRegex.MatchString(new_dns_param.dns_servers_path) {
        fmt.Println("Error: Variable DNS_RESOLVERPATH is wrong check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")
        log.Error("Error: Variable DNS_RESOLVERPATH is wrong check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")
        return false
    }

    new_dns_param.timeout, err = strconv.Atoi(os.Getenv("DNS_TIMEOUT"))
    if err != nil {
        log.Error("Error: Variable DNS_TIMEOUT is empty or wrong in ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.error:", err)
        return false
    }

    delimeter := os.Getenv("DELIMETER")
    if len(delimeter) == 1 {
        new_dns_param.delimeter = rune(delimeter[0])
    } else {
        log.Error("Error: Variable DELIMETER is not a single character. Check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")
        fmt.Println("Error: Variable DELIMETER is not a single character. Check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")

        return false
    }

    new_dns_param.delimeter_for_additional_field = os.Getenv("DELIMETER_FOR_ADDITIONAL_PARAM")
    if len(new_dns_param.delimeter_for_additional_field) < 1 {
        fmt.Println("Error: Variable DELIMETER_FOR_ADDITIONAL_PARAM is wrong check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")
        log.Error("Error: Variable DELIMETER_FOR_ADDITIONAL_PARAM is wrong check ", Config.conf_path, " file. Path:'", new_dns_param.dns_servers_path, "'.")
        return false
    }

    new_dns_param.dns_servers_file_md5hash = Dns_param.dns_servers_file_md5hash

    Dns_param = new_dns_param
    return true
}


func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no local IP address found")
}


func basicAuth() string {
    auth := Prometheus.username + ":" + Prometheus.password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func collectLabels(server Resolver, r_header dns.MsgHdr) []promwrite.Label {
    var label promwrite.Label

    labels := []promwrite.Label{
        {
            Name:  "__name__",
            Value: Prometheus.metric,
        },
        {
            Name: "server",
            Value: server.server,
        },
        {
            Name: "server_ip",
            Value: server.server_ip,
        },
        {
            Name: "domain",
            Value: server.domain,
        },
        {
            Name: "location",
            Value: server.location,
        },
        {
            Name: "site",
            Value: server.site,
        },
        {
            Name: "watcher",
            Value: Config.hostname,
        },
        {
            Name: "watcher_ip",
            Value: Config.ip,
        },
        {
            Name: "watcher_security_zone",
            Value: Config.securityZone,
        },
        {
            Name: "watcher_location",
            Value: Config.location,
        },
        {
            Name: "protocol",
            Value: server.protocol,
        },
        {
            Name: "server_security_zone",
            Value: server.server_security_zone,
        },
        {
            Name: "maintenance_mode",
            Value: strconv.FormatBool(server.maintenance_mode),
        },
    }

    label.Name = "zonename"
    label.Value = server.zonename
    labels = append(labels, label)

    if Prometheus.metrics.authenticatedData {
        label.Name = "authenticated_data"
        label.Value = strconv.FormatBool(r_header.AuthenticatedData)
        labels = append(labels, label)
    }
    if Prometheus.metrics.authoritative {
        label.Name = "authoritative"
        label.Value = strconv.FormatBool(r_header.Authoritative)
        labels = append(labels, label)
    }
    if Prometheus.metrics.checkingDisabled {
        label.Name = "checking_disabled"
        label.Value = strconv.FormatBool(r_header.CheckingDisabled)
        labels = append(labels, label)
    }
    if Prometheus.metrics.opscodes {
        label.Name = "opscodes"
        label.Value = strconv.Itoa(r_header.Opcode)
        labels = append(labels, label)
    }
    if Prometheus.metrics.rcode {
        label.Name = "rcode"
        label.Value = strconv.Itoa(r_header.Rcode)
        labels = append(labels, label)
    }
    if Prometheus.metrics.recursionAvailable {
        label.Name = "recursion_available"
        label.Value = strconv.FormatBool(r_header.RecursionAvailable)
        labels = append(labels, label)
    }
    if Prometheus.metrics.recursionDesired {
        label.Name = "recursion_desired"
        label.Value = strconv.FormatBool(r_header.RecursionDesired)
        labels = append(labels, label)
    }
    if Prometheus.metrics.truncated {
        label.Name = "truncated"
        label.Value = strconv.FormatBool(r_header.Truncated)
        labels = append(labels, label)
    }
    if Prometheus.metrics.polling_rate {
        label.Name = "polling_rate"
        label.Value = strconv.Itoa(server.query_count_rps)
        labels = append(labels, label)
    }
    if Prometheus.metrics.recusrion {
        label.Name = "recursion"
        label.Value = strconv.FormatBool(server.recursion)
        labels = append(labels, label)
    }
    return labels
}


func bufferTimeSeries(server Resolver, tm time.Time, value float64, response_header dns.MsgHdr) {
    Mu.Lock()
	defer Mu.Unlock()
    if len(Buffer) >= Config.buffer_size {
        go sendVM(Buffer)
        Buffer = nil
        return
    }
    instance := promwrite.TimeSeries{
        Labels: collectLabels(server, response_header),
        Sample: promwrite.Sample{
            Time:  tm,
            Value: value,
        },
    }
    Buffer = append(Buffer, instance)
}


func sendVM(items []promwrite.TimeSeries) bool {
    client := promwrite.NewClient(Prometheus.url)
    
    req := &promwrite.WriteRequest{
        TimeSeries: items,
    }
    log.Debug("TimeSeries:", items)
    for i := 0; i < Prometheus.retries; i++ {
        _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth()}))
        if err == nil {
            log.Debug("Remote write to VM succesfull. URL:", Prometheus.url ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
            return true
        }
        log.Warn("Remote write to VM failed. Retry ", i+1, " of ", Prometheus.retries, ". URL:", Prometheus.url, ", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err)
    }
    log.Error("Remote write to VM failed. URL:", Prometheus.url ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
    log.Debug("Request:", req)
    return false
}


func dnsResolve(server Resolver) {
    var host string
    c := dns.Client{Timeout: time.Duration(Dns_param.timeout) * time.Second}
    c.Net = server.protocol
    m := dns.Msg{}
    if server.recursion {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.prefix + "." + server.zonename
    } else {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.prefix + "." + server.zonename
    }
    request_time := time.Now()
    m.SetQuestion(host+".", dns.TypeA)
    r, t, err := c.Exchange(&m, server.server_ip+":53")
    if err != nil {
        log.Debug("Server:", server, ",TC: false", ", host:", host, ", Rcode: 3842, Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", server.query_count_rps, "Recursion:", server.recursion, ", error:", err)
        bufferTimeSeries(server, request_time, float64(t), dns.MsgHdr{ Rcode: 3842})
    } else {
        if len(r.Answer) == 0 {
            log.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", r.MsgHdr.Rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", server.query_count_rps, "Recursion:", server.recursion)
            bufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }  else {
            rcode := r.MsgHdr.Rcode
            if r.Answer[0].(*dns.A).A.To4().String() != "1.1.1.1" {
                rcode = 3841
                r.MsgHdr.Rcode = 3841
            }
            log.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", server.query_count_rps, "Recursion:", server.recursion)
            bufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }
    }
}


func dnsPolling(server Resolver, stop <-chan struct{}) {
    if server.maintenance_mode {
        server.query_count_rps = 1
    }
    for {
        select {
            default:
                go dnsResolve(server)
                time.Sleep(time.Duration(1000 / server.query_count_rps) * time.Millisecond)
            case <-stop:
                return
        }
    }
}


func createPolling() {
    if Polling {
        close(Polling_chan)
        time.Sleep(1 * time.Second)
    }    
    Polling_chan = make(chan struct{})
    Polling = false
    for _, r := range  DnsServers {
        go dnsPolling(r, Polling_chan)
    }
    Polling = true
}


func main() {
    sl := log.GetLevel()
    log.SetLevel(log.InfoLevel)
    log.Info("Frequency DNS cheker start.")
    log.Info("Prometheus info: url:", Prometheus.url , ", auth:", Prometheus.auth, ", username:", Prometheus.username, ", metric_name:", Prometheus.metric)
    log.Info("DNS info: DNS server count:", len(DnsServers) , ", answer_timeout:", Dns_param.timeout)
    log.Info("Debug level:", sl.String() )
    log.SetLevel(sl)

    currentTime := time.Now()
	var startTime = currentTime.Truncate(time.Second).Add(time.Second)
	var duration = startTime.Sub(currentTime)
	time.Sleep(duration)

    ticker := time.NewTicker(time.Duration(Config.check_interval) * time.Minute)
    go func() {
        for range ticker.C {
            checkConfig()
        }
    }()
    
    createPolling()
    
    select {}
}
