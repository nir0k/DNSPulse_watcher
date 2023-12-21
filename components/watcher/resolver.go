package watcher

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

var (
    Polling bool
    Polling_chan chan struct{}
	ResolverConf sqldb.ResolversConfiguration
)

func DnsResolve(server sqldb.Resolver) bool {
    request_time := time.Now()
    if server.ServiceMode{
        // fmt.Println("-")
        log.AppLog.Debug("Server:", server.Server, ",TC: false, host:, Rcode: 0, Protocol:, r_time:", request_time, ", r_duration: 0, polling rate:", server.QueryCount)
        BufferTimeSeries(server, request_time, float64(0), dns.MsgHdr{ Rcode: 0})
        return false
    }
    var host string
    c := dns.Client{Timeout: time.Duration(ResolverConf.PullTimeout) * time.Second}
    c.Net = server.Protocol
    m := dns.Msg{}
    if server.Recursion {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Prefix + "." + server.Zonename
    } else {
        host = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + server.Prefix + "." + server.Zonename
    }
    m.SetQuestion(host+".", dns.TypeA)
    log.AppLog.Debug("m.SetQuestion", m)
    r, t, err := c.Exchange(&m, server.IPAddress+":53")
    if err != nil {
        // fmt.Println("-")
        log.AppLog.Debug("Server:", server, ",TC: false", ", host:", host, ", Rcode: 50, Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.QueryCount, ", Recursion:", server.Recursion, ", error:", err)
        BufferTimeSeries(server, request_time, float64(t), dns.MsgHdr{ Rcode: 50})
    } else {
        if len(r.Answer) == 0 {
            // fmt.Println("-")
            log.AppLog.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", r.MsgHdr.Rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.QueryCount, ", Recursion:", server.Recursion)
            BufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }  else {
            rcode := r.MsgHdr.Rcode
            if r.Answer[0].(*dns.A).A.To4().String() != "1.1.1.1" {
                rcode = 30
                r.MsgHdr.Rcode = 30
            }
            // fmt.Println("-")
            log.AppLog.Debug("Server:", server, ", TC:", r.MsgHdr.Truncated, ", host:", host, ", Rcode:", rcode, ", Protocol:", c.Net, ", r_time:", request_time.Format("2006/01/02 03:04:05.000"), ", r_duration:", t, ", polling rate:", server.QueryCount, ", Recursion:", server.Recursion)
            BufferTimeSeries(server, request_time, float64(t), r.MsgHdr)
        }
    }
    return true
}

func dnsPolling(server sqldb.Resolver, stop <-chan struct{}) {
    if server.ServiceMode {
        server.QueryCount = 1
    }
    for {
        select {
            default:
                go DnsResolve(server)
                time.Sleep(time.Duration(1000 / server.QueryCount) * time.Millisecond)
            case <-stop:
                return
        }
    }
}


func CreatePolling() {
    var err error
	resolvers, err := sqldb.GetResolvers(sqldb.AppDB)
	if err != nil {
		log.AppLog.Error("Failed to get resolvers from db, error:", err)
		return
	}
    // PrometheusConfig, _ = sqldb.GetPrometheusConfig(sqldb.AppDB)
    // PrometheusLabel, _ = sqldb.GetPrometheusLabelConfig(sqldb.AppDB)
    // WatcherConfig, _ = sqldb.GetWatcherConfig(sqldb.AppDB)
    Config, _ = sqldb.GetConfgurations(sqldb.AppDB)
    if Polling {
        close(Polling_chan)
        time.Sleep(1 * time.Second)
    }    
    Polling_chan = make(chan struct{})
    Polling = false
    for _, r := range resolvers {
        go dnsPolling(r, Polling_chan)
    }
    Polling = true
}
