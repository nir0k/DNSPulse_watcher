package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/castai/promwrite"
	"github.com/joho/godotenv"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)


type prometheus struct {
    url string
    auth  bool
    username string
    password string
    metric string
}


type dns_param struct {
    protocol string
    timeout  int
    polling_rate int
    host_postfix string
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
}


var (
    Prometheus prometheus
    Dns_param dns_param
    Log_conf log_conf   
    Config config
    DnsList []string
    Buffer []promwrite.TimeSeries
)


func init() {
    Config.conf_path = ".env"
    state := readConfig()
    if !state {
        fmt.Println("Error load configuration parametrs. check config in .env files")
        log.Fatal("Error load configuration parametrs. check config in .env files")
    }
    state = readDNSServers()
    if !state {
        fmt.Println("Error load dns server list from file '", Dns_param.dns_servers_path, "'. check config in .env files")
        log.Fatal("Error load dns server list from file '", Dns_param.dns_servers_path, "'. check config in .env files")
    }
}


func isValidDNSName(name string) bool {
    dnsNameRegex := regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)$`)
    return dnsNameRegex.MatchString(name)
}


func isValidURL(inputURL string) bool {
    _, err := url.ParseRequestURI(inputURL)
    return err == nil
}


func isAlphaNumericWithDashOrUnderscore(input string) bool {
    validRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
    return validRegex.MatchString(input)
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
    ConfigCheckTicker := time.NewTicker(time.Duration(Config.check_interval) * time.Minute)
    select {
        case <-ConfigCheckTicker.C:
            conf_state, _ := compareFileHash(Config.conf_path, Config.conf_md5hash)
            if !conf_state {
                log.Info("Config has been changed")
                readConfig()
            }

            resolvers_state, _ := compareFileHash(Dns_param.dns_servers_path, Dns_param.dns_servers_file_md5hash)
            if !resolvers_state {
                log.Info("List of DNS service has been changed")
                state := readDNSServers()
                if !state {
                    log.Warn("New List of DNS service is wrong. Use old list of DNS service")
                }
            }
    }
}


func readConfig() bool {
    var (
        new_log_conf log_conf
        new_dns_param dns_param
        new_prometheus prometheus
        new_check_interval int
    )

    err := godotenv.Overload()
    if err != nil {
        fmt.Println("Error loading .env file", err)
        log.Error("Error loading .env file", err)
        return false
    }

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

    new_dns_param.dns_servers_path = os.Getenv("DNS_RESOLVERPATH")
    validRPathRegex := regexp.MustCompile("^[a-zA-Z0-9-_/.]+$")
    if !validRPathRegex.MatchString(new_dns_param.dns_servers_path) {
        fmt.Println("Error: Variable DNS_RESOLVERPATH is wrong check .env file. Path:'", new_dns_param.dns_servers_path, "'.")
        log.Error("Error: Variable DNS_RESOLVERPATH is wrong check .env file. Path:'", new_dns_param.dns_servers_path, "'.")
        return false
    }

    new_dns_param.host_postfix = os.Getenv("DNS_HOSTPOSTFIX")    
    if !isValidDNSName(new_dns_param.host_postfix) {
        log.Error("Error: Variable HOSTPOSTFIX is wrong in .env file.")
        return false
    }
    new_dns_param.polling_rate, err = strconv.Atoi(os.Getenv("DNS_POLLING_RATE"))
    if err != nil {
        log.Error("Warning: Variable DNS_POLLING_RATE is empty or wrong in .env file. error:", err)
        return false
    }

    new_dns_param.timeout, err = strconv.Atoi(os.Getenv("DNS_TIMEOUT"))
    if err != nil {
        log.Error("Warning: Variable DNS_TIMEOUT is empty or wrong in .env file. error:", err)
        return false
    }

    new_dns_param.protocol = os.Getenv("DNS_PROTOCOL")
    regexpPattern, err := regexp.Compile("^(udp[4,6]|tcp[4,6]|tcp|udp)$")
    if err != nil {
        log.Error("Error compiling regex. err:", err)
        return false
    } else if !regexpPattern.MatchString(new_dns_param.protocol) {
        log.Error("Error: Variable DNS_PROTOCOL is empty or wrong in .env file.")
        return false
    }

    new_prometheus.url = os.Getenv("PROM_URL")
    if !isValidURL(new_prometheus.url) {
        log.Error("Error: Variable PROM_URL is required in .env file. Please add this variable with value")
        return false
    }

    new_prometheus.metric = os.Getenv("PROM_METRIC")
    if !isAlphaNumericWithDashOrUnderscore(new_prometheus.metric) {
        log.Error("Error: Variable PROM_METRIC is empty or wrong in .env file.")
        return false
    }

    new_prometheus.auth, err = strconv.ParseBool(os.Getenv("PROM_AUTH"))
    if err != nil {
        log.Error("Error: Variable PROM_AUTH is empty or wrong in .env file. error:", err)
        return false
    }

    if new_prometheus.auth {
        new_prometheus.username = os.Getenv("PROM_USER")
        if len(new_prometheus.username) == 0 {
            log.Error("Error: Variable PROM_USER is required in .env file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
        new_prometheus.password = os.Getenv("PROM_PASS")
        if len(new_prometheus.password) == 0 {
            log.Error("Error: Variable PROM_PASS is required in .env file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
            return false
        }
    }

    new_check_interval, err = strconv.Atoi(os.Getenv("CONF_CHECK_INTERVAL"))
    if err != nil {
        log.Error("Warning: Variable CONF_CHECK_INTERVAL is empty or wrong in .env file. error:", err)
        return false
    }

    Config.buffer_size, err = strconv.Atoi(os.Getenv("BUFFER_SIZE"))
    if err != nil {
        log.Error("Warning: Variable BUFFER_SIZE is empty or wrong in .env file. error:", err)
        return false
    }

    Config.conf_md5hash, err = calculateHash(Config.conf_path, md5.New)
    if err != nil {
        log.Error("Error: calculate hash to file '", Config.conf_path, "'")
        return false
    }

    Dns_param.dns_servers_file_md5hash, err = calculateHash(Config.conf_path, md5.New)
    if err != nil {
        log.Error("Error: calculate hash to file '", Config.conf_path, "'")
        return false
    }

    Log_conf = new_log_conf
    Dns_param = new_dns_param
    Prometheus = new_prometheus
    Config.check_interval = new_check_interval
    
    return true
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


func readDNSServers() bool {
    file, err := os.Open(Dns_param.dns_servers_path)
    if err != nil {
        log.Error("Error read file ", Dns_param.dns_servers_path,"| error: ", err)
        return false
    }
    defer file.Close()
    var dnslist []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        dnslist = append(dnslist, scanner.Text())
    }
    if len(dnslist) == 0 {
        log.Error("Error: File ", Dns_param.dns_servers_path, " is empty")
        return false
    }
    DnsList = dnslist
    log.Info("DNS info: DNS server count:", len(DnsList))
    return true
}


func basicAuth() string {
    auth := Prometheus.username + ":" + Prometheus.password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func bufferTimeSeries(server string, tc bool, Rcode int, protocol string, tm time.Time, value float64) {
    if len(Buffer) >= Config.buffer_size {
        go sendVM(Buffer)
        Buffer = []promwrite.TimeSeries{}
        return
    }
    instance := promwrite.TimeSeries{
        Labels: []promwrite.Label{
            {
                Name:  "__name__",
                Value: Prometheus.metric,
            },
            {
                Name: "server",
                Value: server,
            },
            {
                Name: "Truncated",
                Value: strconv.FormatBool(tc),
            },
            {
                Name: "Rcode",
                Value: strconv.Itoa(Rcode),
            },
            {
                Name: "Protocol",
                Value: protocol,
            },
        },
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
    for i := 0; i <= 3; i++ {
        _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth()}))
        if err == nil {
            log.Debug("Remote write to VM succesfull. URL:", Prometheus.url ,", Username:", Prometheus.username,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
            return true
        }
        if i > 0 {
            log.Warn("Remote write to VM succesfull. Retry ", i+1, " of 3. URL:", Prometheus.url ,", Username:", Prometheus.username,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err, ", request:", req)
        }
    }
    return false
}


func dnsResolve(target string, server string, server_label string) {
    c := dns.Client{Timeout: time.Duration(Dns_param.timeout) * time.Second}
    c.Net = Dns_param.protocol
    m := dns.Msg{}
    host := strconv.FormatInt(time.Now().UnixNano(), 10) + "." + target
    request_time := time.Now()
    m.SetQuestion(host+".", dns.TypeA)
    r, t, err := c.Exchange(&m, server+":53")
    if err != nil {
        log.Debug("Server:", server_label, ",TC: false", ", host:", host, ", Rcode: 3842, Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", error:", err)
        bufferTimeSeries(server_label, false, 3842, c.Net, request_time, float64(t))
    } else {
        if len(r.Answer) == 0 {
            log.Debug("Server:", server_label, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", r.MsgHdr.Rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t)
            bufferTimeSeries(server_label, r.MsgHdr.Truncated, r.MsgHdr.Rcode, c.Net, request_time, float64(t))
        } else {
            rcode := r.MsgHdr.Rcode
            if r.Answer[0].(*dns.A).A.To4().String() != "1.1.1.1" {
                rcode = 3841
            }
            log.Debug("Server:", server_label, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t)
            bufferTimeSeries(server_label, r.MsgHdr.Truncated, r.MsgHdr.Rcode, c.Net, request_time, float64(t))
        }
    }
}


func main() {
    log.Info("Frequency DNS cheker start.")
    log.Info("Prometheus info: url:", Prometheus.url , ", auth:", Prometheus.auth, ", username:", Prometheus.username, ", metric_name:", Prometheus.metric)
    log.Info("DNS info: DNS server count:", len(DnsList) , ", hostname_postfix:", Dns_param.host_postfix, ", answer_timeout:", Dns_param.timeout, ", polling_rate:", Dns_param.polling_rate)
    
    currentTime := time.Now()
	var startTime = currentTime.Truncate(time.Second).Add(time.Second)
	var duration = startTime.Sub(currentTime)
	time.Sleep(duration)
    
    go checkConfig()
    i := 0
    for {
        i = 0
        for _, r := range DnsList {
            i++
            go dnsResolve(Dns_param.host_postfix, r, fmt.Sprintf("%s%d", r, i))
        }
        time.Sleep(time.Duration(Dns_param.polling_rate) * time.Millisecond)
    }
}
