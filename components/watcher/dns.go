package watcher

import (
	"strconv"
	"time"
	"github.com/miekg/dns"
	// log "github.com/sirupsen/logrus"
    "github.com/nir0k/HighFrequencyDNSChecker/components/log"
)

var (
    Polling bool
    Polling_chan chan struct{}
)

func DnsResolve(server Resolver) bool {
    request_time := time.Now()
    if server.Service_mode{
        log.AppLog.Debug("Server:", server, ",TC: false, host:, Rcode: 0, Protocol:, r_time:", request_time, ", r_duration: 0, polling rate:", server.Query_count_rps, ", Recursion:")
        bufferTimeSeries(server, request_time, float64(0), dns.MsgHdr{ Rcode: 0})
        return false
    }
    var host string
    c := dns.Client{Timeout: time.Duration(Dns_param.Timeout) * time.Second}
    c.Net = server.Protocol
    m := dns.Msg{}
    if server.Recursion {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Prefix + "." + server.Zonename
    } else {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Prefix + "." + server.Zonename
    }
    m.SetQuestion(host+".", dns.TypeA)
    log.AppLog.Debug("m.SetQuestion", m)
    r, t, err := c.Exchange(&m, server.Server_ip+":53")
    if err != nil {
        log.AppLog.Debug("Server:", server, ",TC: false", ", host:", host, ", Rcode: 50, Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.Query_count_rps, ", Recursion:", server.Recursion, ", error:", err)
        bufferTimeSeries(server, request_time, float64(t), dns.MsgHdr{ Rcode: 50})
    } else {
        if len(r.Answer) == 0 {
            log.AppLog.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", r.MsgHdr.Rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.Query_count_rps, ", Recursion:", server.Recursion)
            bufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }  else {
            rcode := r.MsgHdr.Rcode
            if r.Answer[0].(*dns.A).A.To4().String() != "1.1.1.1" {
                rcode = 30
                r.MsgHdr.Rcode = 30
            }
            log.AppLog.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.Query_count_rps, ", Recursion:", server.Recursion)
            bufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }
    }
    return true
}

func dnsPolling(server Resolver, stop <-chan struct{}) {
    if server.Service_mode {
        server.Query_count_rps = 1
    }
    for {
        select {
            default:
                go DnsResolve(server)
                time.Sleep(time.Duration(1000 / server.Query_count_rps) * time.Millisecond)
            case <-stop:
                return
        }
    }
}


func CreatePolling() {
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