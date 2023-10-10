package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/castai/promwrite"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Resolver struct {
	Server			string `csv:"server"`
	Server_ip		string `csv:"server_ip"`
	Domain			string `csv:"domain"`
    Location        string `csv:"location"`
    Site            string `csv:"site"`
    Suffix          string `csv:"suffix"`
    Protocol        string `csv:"protocol"`
    Zonename		string `csv:"zonename"`
    Query_count_rps string `csv:"query_count_rps"`
    Zonename_with_recursion string `csv:"zonename_with_recursion"`
    Query_count_with_recursion string `csv:"query_count_with_recursion_rps"`
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


func readDNSServersFromCSV() bool {
    dns_list := []Resolver{}
    clientsFile, err := os.OpenFile(Dns_param.dns_servers_path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error("Error read file ", Dns_param.dns_servers_path,"| error: ", err)
        return false
	}
	defer clientsFile.Close()
	if err := gocsv.UnmarshalFile(clientsFile, &dns_list); err != nil {
		log.Error("Error Unmarshal file ", Dns_param.dns_servers_path,"| error: ", err)
        return false
	}
    DnsServers = dns_list
    
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
        }
    }

    resolvers_state, _ := compareFileHash(Dns_param.dns_servers_path, Dns_param.dns_servers_file_md5hash)
    if !resolvers_state {
        sl := log.GetLevel()
        log.Info("List of DNS service has been changed")
        log.SetLevel(sl)
        state := readDNSServersFromCSV()
        if !state {
            log.Warn("New List of DNS service is wrong. Use old list of DNS service")
        } else {
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
        log.Error("Error: Variable DNS_TIMEOUT is empty or wrong in ", Config.conf_path, " file. error:", err)
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
	return "", fmt.Errorf("No local IP address found")
}


func basicAuth() string {
    auth := Prometheus.username + ":" + Prometheus.password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func collectLabels(server Resolver, recursion bool, r_header dns.MsgHdr, polling_rate int) []promwrite.Label {
    var label promwrite.Label

    labels := []promwrite.Label{
        {
            Name:  "__name__",
            Value: Prometheus.metric,
        },
        {
            Name: "server",
            Value: server.Server,
        },
        {
            Name: "server_ip",
            Value: server.Server_ip,
        },
        {
            Name: "domain",
            Value: server.Domain,
        },
        {
            Name: "location",
            Value: server.Location,
        },
        {
            Name: "site",
            Value: server.Site,
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
            Value: server.Protocol,
        },
    }

    label.Name = "zonename"
    if recursion {
        label.Value = server.Zonename_with_recursion
    } else {
        label.Value = server.Zonename
    }
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
        if recursion {
            label.Value = server.Query_count_with_recursion
        } else {
            label.Value = server.Query_count_rps
        }
        labels = append(labels, label)
    }
    if Prometheus.metrics.recusrion {
        label.Name = "recursion"
        label.Value = strconv.FormatBool(recursion)
        labels = append(labels, label)
    }
    return labels
}


func bufferTimeSeries(server Resolver, tm time.Time, value float64, recursion bool, response_header dns.MsgHdr, polling_rate int) {
    Mu.Lock()
	defer Mu.Unlock()
    if len(Buffer) >= Config.buffer_size {
        go sendVM(Buffer)
        Buffer = nil
        return
    }
    instance := promwrite.TimeSeries{
        Labels: collectLabels(server, recursion, response_header, polling_rate),
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


func dnsResolve(server Resolver, recursion bool, polling_rate int) {
    var host string
    c := dns.Client{Timeout: time.Duration(Dns_param.timeout) * time.Millisecond}
    c.Net = server.Protocol
    m := dns.Msg{}
    if recursion {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Suffix + "." + server.Zonename_with_recursion
    } else {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Suffix + "." + server.Zonename
    }
    request_time := time.Now()
    m.SetQuestion(host+".", dns.TypeA)
    r, t, err := c.Exchange(&m, server.Server_ip+":53")
    if err != nil {
        log.Debug("Server:", server, ",TC: false", ", host:", host, ", Rcode: 3842, Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", polling_rate, "Recursion:", recursion, ", error:", err)
        bufferTimeSeries(server, request_time, float64(t), recursion, dns.MsgHdr{ Rcode: 3842}, polling_rate)
    } else {
        if len(r.Answer) == 0 {
            log.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", r.MsgHdr.Rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", polling_rate, "Recursion:", recursion)
            bufferTimeSeries(server, request_time, float64(t), recursion, r.MsgHdr, polling_rate)
        }  else {
            rcode := r.MsgHdr.Rcode
            if r.Answer[0].(*dns.A).A.To4().String() != "1.1.1.1" {
                rcode = 3841
            }
            log.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, "polling rate:", polling_rate, "Recursion:", recursion)
            bufferTimeSeries(server, request_time, float64(t), recursion, r.MsgHdr, polling_rate)
        }
    }
}


func dnsPolling(server Resolver, recursion bool, stop <-chan struct{}) { 
    if recursion {
        polling_rate, _ := strconv.Atoi(server.Query_count_with_recursion)
        for {
            select {
                default:
                    go dnsResolve(server, true, polling_rate)
                    time.Sleep(time.Duration(1000 / polling_rate) * time.Millisecond)
                case <-stop:
                    return
            }
        }
    } else {
        polling_rate, _ := strconv.Atoi(server.Query_count_rps)
        for {
            select {
                default:
                    go dnsResolve(server, false, polling_rate)
                    time.Sleep(time.Duration(1000 / polling_rate) * time.Millisecond)
                case <-stop:
                    return
            }
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
        if r.Zonename != "" {
            go dnsPolling(r, false, Polling_chan)
        }
        if r.Zonename_with_recursion != "" {
            go dnsPolling(r, true, Polling_chan)
        }
    }
    Polling = true
}


func main() {
    sl := log.GetLevel()
    log.SetLevel(log.InfoLevel)
    log.Info("Frequency DNS cheker start.")
    log.Info("Prometheus info: url:", Prometheus.url , ", auth:", Prometheus.auth, ", username:", Prometheus.username, ", metric_name:", Prometheus.metric)
    log.Info("DNS info: DNS server count:", len(DnsServers) , ", answer_timeout:", Dns_param.timeout)
    log.SetLevel(sl)

    currentTime := time.Now()
	var startTime = currentTime.Truncate(time.Second).Add(time.Second)
	var duration = startTime.Sub(currentTime)
	time.Sleep(duration)

    ticker := time.NewTicker(time.Duration(Config.check_interval) * time.Minute)
    go func() {
        for {
           select {
            case <- ticker.C:
                checkConfig()
            }
        }
    }()
    
    createPolling()
    
    select {}
}
