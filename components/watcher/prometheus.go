package watcher

import (
	"context"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/castai/promwrite"
	"github.com/miekg/dns"
	// log "github.com/sirupsen/logrus"
    "github.com/nir0k/HighFrequencyDNSChecker/components/log"
)

var (
    Prometheus prometheus
    Buffer []promwrite.TimeSeries
)


func basicAuth() string {
    auth := Prometheus.Username + ":" + Prometheus.Password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func collectLabels(server Resolver, r_header dns.MsgHdr) []promwrite.Label {
    var label promwrite.Label

    labels := []promwrite.Label{
        {
            Name:  "__name__",
            Value: Prometheus.Metric,
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
            Value: Config.Hostname,
        },
        {
            Name: "watcher_ip",
            Value: Config.Ip,
        },
        {
            Name: "watcher_security_zone",
            Value: Config.SecurityZone,
        },
        {
            Name: "watcher_location",
            Value: Config.Location,
        },
        {
            Name: "protocol",
            Value: server.Protocol,
        },
        {
            Name: "server_security_zone",
            Value: server.Server_security_zone,
        },
        {
            Name: "service_mode",
            Value: strconv.FormatBool(server.Service_mode),
        },
    }

    label.Name = "zonename"
    label.Value = server.Zonename
    labels = append(labels, label)

    if Prometheus.Metrics.AuthenticatedData {
        label.Name = "authenticated_data"
        label.Value = strconv.FormatBool(r_header.AuthenticatedData)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Authoritative {
        label.Name = "authoritative"
        label.Value = strconv.FormatBool(r_header.Authoritative)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.CheckingDisabled {
        label.Name = "checking_disabled"
        label.Value = strconv.FormatBool(r_header.CheckingDisabled)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Opscodes {
        label.Name = "opscodes"
        label.Value = strconv.Itoa(r_header.Opcode)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Rcode {
        label.Name = "rcode"
        label.Value = strconv.Itoa(r_header.Rcode)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.RecursionAvailable {
        label.Name = "recursion_available"
        label.Value = strconv.FormatBool(r_header.RecursionAvailable)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.RecursionDesired {
        label.Name = "recursion_desired"
        label.Value = strconv.FormatBool(r_header.RecursionDesired)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Truncated {
        label.Name = "truncated"
        label.Value = strconv.FormatBool(r_header.Truncated)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Polling_rate {
        label.Name = "polling_rate"
        label.Value = strconv.Itoa(server.Query_count_rps)
        labels = append(labels, label)
    }
    if Prometheus.Metrics.Recusrion {
        label.Name = "recursion"
        label.Value = strconv.FormatBool(server.Recursion)
        labels = append(labels, label)
    }
    return labels
}


func bufferTimeSeries(server Resolver, tm time.Time, value float64, response_header dns.MsgHdr) {
    Mu.Lock()
	defer Mu.Unlock()
    if len(Buffer) >= Config.Buffer_size {
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
    client := promwrite.NewClient(Prometheus.Url)
    
    req := &promwrite.WriteRequest{
        TimeSeries: items,
    }
    log.AppLog.Debug("TimeSeries:", items)
    for i := 0; i < Prometheus.Retries; i++ {
        _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth()}))
        if err == nil {
            log.AppLog.Debug("Remote write to VM succesfull. URL:", Prometheus.Url ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
            return true
        }
        log.AppLog.Warn("Remote write to VM failed. Retry ", i+1, " of ", Prometheus.Retries, ". URL:", Prometheus.Url, ", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err)
    }
    log.AppLog.Error("Remote write to VM failed. URL:", Prometheus.Url ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
    log.AppLog.Debug("Request:", req)
    return false
}