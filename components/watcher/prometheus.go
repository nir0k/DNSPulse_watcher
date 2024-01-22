package polling

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"context"
	"encoding/base64"
	"strconv"
	"sync"
	"time"

	"github.com/castai/promwrite"
	"github.com/miekg/dns"
)

var (
    Buffer []promwrite.TimeSeries
    Mu sync.Mutex
)


func basicAuth(conf datastore.PrometheusConfStruct) string {
    auth := conf.Username + ":" + conf.Password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}


func collectLabels(server datastore.PollingHost, r_header dns.MsgHdr, conf datastore.PrometheusConfStruct, confG datastore.GeneralConfigStruct) []promwrite.Label {
    promLabels := datastore.GetConfig().Prometheus.Labels
    var label promwrite.Label

    labels := []promwrite.Label{
        {
            Name:  "__name__",
            Value: conf.MetricName,
        },
        {
            Name: "server",
            Value: server.Hostname,
        },
        {
            Name: "server_ip",
            Value: server.IPAddress,
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
            Value: confG.Hostname,
        },
        {
            Name: "watcher_ip",
            Value: confG.IPAddress,
        },
        {
            Name: "watcher_security_zone",
            Value: confG.SecurityZone,
        },
        {
            Name: "watcher_location",
            Value: confG.Location,
        },
        {
            Name: "protocol",
            Value: server.Protocol,
        },
        {
            Name: "server_security_zone",
            Value: server.SecurityZone,
        },
        {
            Name: "service_mode",
            Value: strconv.FormatBool(server.ServiceMode),
        },
    }

    label.Name = "zonename"
    label.Value = server.Zonename
    labels = append(labels, label)

    if promLabels.AuthenticatedData {
        label.Name = "authenticated_data"
        label.Value = strconv.FormatBool(r_header.AuthenticatedData)
        labels = append(labels, label)
    }
    if promLabels.Authoritative {
        label.Name = "authoritative"
        label.Value = strconv.FormatBool(r_header.Authoritative)
        labels = append(labels, label)
    }
    if promLabels.CheckingDisabled {
        label.Name = "checking_disabled"
        label.Value = strconv.FormatBool(r_header.CheckingDisabled)
        labels = append(labels, label)
    }
    if promLabels.Opcode {
        label.Name = "opscodes"
        label.Value = strconv.Itoa(r_header.Opcode)
        labels = append(labels, label)
    }
    if promLabels.Rcode {
        label.Name = "rcode"
        label.Value = strconv.Itoa(r_header.Rcode)
        labels = append(labels, label)
    }
    if promLabels.RecursionAvailable {
        label.Name = "recursion_available"
        label.Value = strconv.FormatBool(r_header.RecursionAvailable)
        labels = append(labels, label)
    }
    if promLabels.RecursionDesired {
        label.Name = "recursion_desired"
        label.Value = strconv.FormatBool(r_header.RecursionDesired)
        labels = append(labels, label)
    }
    if promLabels.Truncated {
        label.Name = "truncated"
        label.Value = strconv.FormatBool(r_header.Truncated)
        labels = append(labels, label)
    }
    if promLabels.PollingRate {
        label.Name = "polling_rate"
        label.Value = strconv.Itoa(server.QueryCount)
        labels = append(labels, label)
    }
    if promLabels.Recursion {
        label.Name = "recursion"
        label.Value = strconv.FormatBool(server.Recursion)
        labels = append(labels, label)
    }
    return labels
}


func BufferTimeSeries(server datastore.PollingHost, tm time.Time, value float64, response_header dns.MsgHdr) {
    conf := datastore.GetConfig().Prometheus
    confGeneral := datastore.GetConfig().General
    Mu.Lock()
	defer Mu.Unlock()
    if len(Buffer) >= conf.BufferSize {
        go sendVM(Buffer, conf)
        Buffer = nil
        return
    }
    instance := promwrite.TimeSeries{
        Labels: collectLabels(server, response_header, conf, confGeneral),
        Sample: promwrite.Sample{
            Time:  tm,
            Value: value,
        },
    }
    Buffer = append(Buffer, instance)
}


func sendVM(items []promwrite.TimeSeries, conf datastore.PrometheusConfStruct) bool {
    client := promwrite.NewClient(conf.URL)
    
    req := &promwrite.WriteRequest{
        TimeSeries: items,
    }
    logger.Logger.Debug("TimeSeries:", items)
    for i := 0; i < conf.RetriesCount; i++ {
        _, err := client.Write(context.Background(), req, promwrite.WriteHeaders(map[string]string{"Authorization": "Basic " + basicAuth(conf)}))
        if err == nil {
            logger.Logger.Debug("Remote write to VM succesfull. URL:", conf.URL ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
            return true
        }
        logger.Logger.Warn("Remote write to VM failed. Retry ", i+1, " of ", conf.RetriesCount, ". URL:", conf.URL, ", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"), ", error:", err)
    }
    logger.Logger.Error("Remote write to VM failed. URL:", conf.URL ,", timestamp:", time.Now().Format("2006/01/02 03:04:05.000"))
    logger.Logger.Debug("Request:", req)
    return false
}
