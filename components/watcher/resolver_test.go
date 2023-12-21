package watcher

import (
	"io"
	"net"
	"testing"
	"time"

	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// mockDNSServer starts a mock DNS server that responds with the provided dns.Msg.
func mockDNSServer(response *dns.Msg, t *testing.T, errChan chan error) *dns.Server {
    handler := func(w dns.ResponseWriter, r *dns.Msg) {
        m := response.Copy()
        m.SetReply(r)
        w.WriteMsg(m)
    }

    server := &dns.Server{Addr: "127.0.0.1:0", Net: "udp"}
    dns.HandleFunc("test.", handler)
    go func() {
        if err := server.ListenAndServe(); err != nil {
            errChan <- err
        }
    }()

    return server
}

func TestDnsResolve(t *testing.T) {
	errChan := make(chan error, 1)
    defer close(errChan)

	log.AppLog = logrus.New()
    log.AppLog.Out = io.Discard
    // Mock DNS response
    mockResponse := new(dns.Msg)
    mockResponse.SetReply(&dns.Msg{})
    mockResponse.Answer = append(mockResponse.Answer, &dns.A{
        Hdr: dns.RR_Header{Name: "test.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
        A:   net.ParseIP("127.0.0.1"),
    })

    // Start mock DNS server
    server := mockDNSServer(mockResponse, t, errChan)
    defer server.Shutdown()

	select {
    case err := <-errChan:
        t.Fatalf("Failed to start mock DNS server: %v", err)
    case <-time.After(time.Second): // Adjust timeout as needed
        // No error, continue
    }

    // Get the mock server address
    addr := server.PacketConn.LocalAddr().String()

    // Set up test resolver to query the mock server
    testResolver := sqldb.Resolver{
        Server:    "mockserver",
        IPAddress: addr, // Use mock server address
        Protocol:  "udp",
        // Other necessary fields...
    }

    // Call the DnsResolve function
    result := DnsResolve(testResolver)

    // Assertions
    if !result {
        t.Errorf("DnsResolve failed, expected success")
    }

    // Additional assertions based on mockResponse and expected behavior
}