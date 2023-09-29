package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/castai/promwrite"
	"github.com/joho/godotenv"
	"github.com/miekg/dns"
    log "github.com/sirupsen/logrus"
)


var (
    Resolvers_File_Path string
    HostPrefix string
    TimeDelay int
    Dns_timeout int
    Prometheus_url string
    Prometheus_auth bool
    Prometheus_username string
    Prometheus_password string
    Prometheus_metric_name string
    Debug bool
    Log_file string
    Log_level string
)


func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file", err)
    }
    Log_file = os.Getenv("LOG_FILE")
    Log_level = os.Getenv("LOG_LEVEL")
    file, err := os.OpenFile(Log_file, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
    if err != nil {
        fmt.Println("Error opening file '", Log_file,"': %v", err)
    }

    log.SetFormatter(&log.JSONFormatter{})
    log.SetOutput(file)
    switch Log_level{
        case "debug": log.SetLevel(log.DebugLevel)
        case "info": log.SetLevel(log.InfoLevel)
        case "warning": log.SetLevel(log.WarnLevel)
        case "error": log.SetLevel(log.ErrorLevel)
        case "fatal": log.SetLevel(log.FatalLevel)
        default: log.SetLevel(log.WarnLevel) 
    }
    Resolvers_File_Path = os.Getenv("RESOLVERPATH")
    HostPrefix = os.Getenv("HOSTPREFIX")
    if len(HostPrefix) == 0 {
        log.Fatal("Error: Variable HOSTPREFIX is required in .env file. Please add this variable with value")
    }
    TimeDelay, err = strconv.Atoi(os.Getenv("TIMEDELAY"))
    if err != nil {
        log.Warn("Warning: Variable TIMEDELAY is empty or wrong in .env file. The value equals to 150 will be used. error:", err)
    }
    Prometheus_url = os.Getenv("PROM_URL")
    if len(Prometheus_url) == 0 {
        log.Fatal("Error: Variable PROM_URL is required in .env file. Please add this variable with value")
    }
    Prometheus_metric_name = os.Getenv("PROM_METRIC")
    if len(Prometheus_metric_name) == 0 {
        log.Warn("Warning: Variable PROM_AUTH is empty or wrong in .env file. The value equals to 'dns_resolve' will be used.")
    }
    Prometheus_auth, err = strconv.ParseBool(os.Getenv("PROM_AUTH"))
    if err != nil {
        log.Warn("Warning: Variable PROM_AUTH is empty or wrong in .env file. The value equals to 'false' will be used. error:", err)
    }
    if Prometheus_auth {
        Prometheus_username = os.Getenv("PROM_USER")
        if len(Prometheus_username) == 0 {
            log.Fatal("Error: Variable PROM_USER is required in .env file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
        }
        Prometheus_password = os.Getenv("PROM_PASS")
        if len(Prometheus_password) == 0 {
            log.Fatal("Error: Variable PROM_PASS is required in .env file or variable PROM_AUTH must equals to 'false'. Please add this variable with value")
        }
    }
    Dns_timeout, err = strconv.Atoi(os.Getenv("DNSTIMEOUT"))
    if err != nil {
        log.Warn("Warning: Variable DNSTIMEOUT is empty or wrong in .env file. The value equals to 1 will be used. error:", err)
    }
}


func readLines() ([]string, error) {
    file, err := os.Open(Resolvers_File_Path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}


func basicAuth() string {
    auth := Prometheus_username + ":" + Prometheus_password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func sendVM(server string, tc bool, r_code int, tm time.Time, value float64) bool {
    client := promwrite.NewClient(Prometheus_url)
    req := &promwrite.WriteRequest{
        TimeSeries: []promwrite.TimeSeries{
            {
                Labels: []promwrite.Label{
                    {
                        Name:  "__name__",
                        Value: Prometheus_metric_name,
                    },
                    {
                        Name: "server",
                        Value: server,
                    },
                    {
                        Name: "type",
                        Value: "A",
                    },
                    {
                        Name: "Truncated",
                        Value: strconv.FormatBool(tc),
                    },
                    {
                        Name: "r_code",
                        Value: strconv.Itoa(r_code),
                    },
                },
                Sample: promwrite.Sample{
                    Time:  tm,
                    Value: value,
                },
            },
        },
    }
    _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth()}))
    if err != nil {
        err_count := 0
        for i := 1; i <= 3; i++ {
            err_count = i
            _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth()}))
            if err == nil {
                break
            }
        }
        if err_count < 3 {
            log.Warn("Remote write to VM succesfull. Retry ", err_count, " of 3. URL:", Prometheus_url ,", Username:", Prometheus_username,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err, ", request:", req)
            return true
        }else {
            log.Error("Remote write to VM false. URL:", Prometheus_url ,", Username:", Prometheus_username,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err, ", request:", req)
            return false
        }
    }
    log.Debug("Remote write to VM succesfull. URL:", Prometheus_url ,", Username:", Prometheus_username,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
    return true
}


func dnsResolve(target string, server string, server_label string) {
    c := dns.Client{Timeout: time.Duration(Dns_timeout) * time.Second}
    m := dns.Msg{}
    host := strconv.FormatInt(time.Now().UnixNano(), 10) + "." + target
    request_time := time.Now()
    m.SetQuestion(host+".", dns.TypeA)
    r, t, err := c.Exchange(&m, server+":53")
    if err != nil {
        log.Debug("Server:", server_label, ",TC: false", ", host:", host, ", r_code: 3 , r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", error:", err)
        sendVM(server_label, false, 3, request_time, float64(t))
    } else {
        if len(r.Answer) == 0 {
            log.Debug("Server:", server_label, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", r_code: 2, r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t)
            sendVM(server_label, r.MsgHdr.Truncated, 2, request_time, float64(t))
        } else {
            r_code := 0
            if r.Answer[0].(*dns.A).A.To4().String() == "1.1.1.1" {
                r_code = 1
            }
            log.Debug("Server:", server_label, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", r_code:", r_code,", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t)
            sendVM(server_label, r.MsgHdr.Truncated, r_code, request_time, float64(t))
        }
    }
}

func main() {
    log.Info("Frequency DNS cheker start.")
    log.Info("Prometheus info: url:", Prometheus_url , ", auth:", Prometheus_auth, ", username:", Prometheus_username, ", metric_name:", Prometheus_metric_name)
    resolvers, err := readLines()
    if err != nil {
        log.Fatal("Error read file ", Resolvers_File_Path,"| error: ", err)
    }
    if len(resolvers) == 0 {
        log.Fatal("Error: File ", Resolvers_File_Path, " is empty")
    }
    log.Info("DNS info: DNS server count:", len(resolvers) , ", hostname_prefix:", HostPrefix, ", answer_timeout:", Dns_timeout, ", polling_frequency:", TimeDelay)
    i := 0
    for {
        i = 0
        for _, r := range resolvers {
            i++
            go dnsResolve(HostPrefix, r, fmt.Sprintf("%s%d", r, i))
        }
    }
}
