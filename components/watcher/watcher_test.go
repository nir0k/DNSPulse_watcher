package watcher

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"HighFrequencyDNSChecker/components/log"
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"database/sql"
)

func setupInMemoryDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open in-memory db: %v", err)
    }
    if err = sqldb.InitDB(db); err != nil {
        t.Fatalf("Failed to initialize db: %v", err)
    }
	sqldb.AppDB = db
    return db
}

func insertTestConfigs(t *testing.T, db *sql.DB) {
    // Insert test data for WatcherConfiguration
    _, err := db.Exec("INSERT INTO config_watcher (Location, SecurityZone) VALUES (?, ?)", "TestLocation", "TestSecurityZone")
    if err != nil {
        t.Fatalf("Failed to insert test watcher config: %v", err)
    }

    // Insert test data for MainConfiguration
    _, err = db.Exec(
        "INSERT INTO config_main (HostName, IPAddress, DBname, ConfPath, Sync, UpdateInterval, Hash, LastCheck, LastUpdate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
        "testHost", "192.168.1.1", "testDB", "/test/path", true, 30, "testhash", 1625076000, 1625076000,
    )
    if err != nil {
        t.Fatalf("Failed to insert test main config: %v", err)
    }

    // Insert test data for PrometheusConfiguration
    _, err = db.Exec("INSERT INTO config_prometheus (Url, MetricName, Auth, Username, Password, RetriesCount) VALUES (?, ?, ?, ?, ?, ?)", "http://example.com", "testMetric", 1, "user", "pass", 3)
    if err != nil {
        t.Fatalf("Failed to insert test prometheus config: %v", err)
    }

    // Insert test data for PrometheusLabelConfiguration
    // Assuming PrometheusLabelConfiguration has fields like `Opcode`, `Authoritative`, etc. as boolean.
	insertLabels := []struct {
		Label   string
		Enabled bool
	}{
		{"opcode", false},
		{"authoritative", false},
		{"truncated", true},
		{"rcode", true},
		{"recursionDesired", false},
		{"recursionAvailable", false},
		{"authenticatedData", false},
		{"checkingDisabled", false},
		{"pollingRate", false},
		{"recursion", true},
	}
    for _, l := range insertLabels {
		_, err := db.Exec("INSERT INTO prometheus_label_config (label, isEnable) VALUES (?, ?)", l.Label, l.Enabled)
		if err != nil {
			t.Fatalf("Failed to insert label config for %s: %v", l.Label, err)
		}
	}
}


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
