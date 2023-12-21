package watcher

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"HighFrequencyDNSChecker/components/log"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestDnsResolveServiceMode(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()
    
    insertTestConfigs(t, db)

    log.InitAppLogger()

    resolver := sqldb.Resolver{
        Server: "test_server",
        IPAddress: "1.2.3.4",
        ServiceMode: true,
    }

    performed := DnsResolve(resolver)

    if performed {
        t.Errorf("dnsResolve should not perform resolution when service_mode is true")
    }
}
