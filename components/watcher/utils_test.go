package watcher

import (
	// "HighFrequencyDNSChecker/components/log"
	sqldb "HighFrequencyDNSChecker/components/db"
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

func TestGetPrometheusConfigsFromDB(t *testing.T) {
    db := setupInMemoryDB(t)
    defer db.Close()

    // Insert test data
    insertTestConfigs(t, db)

    // Call the function
    err := GetPrometheusConfigsFromDB()
    if err != nil {
        t.Errorf("Failed to get Prometheus configs: %v", err)
    }

    // Verify results
    // You should replace these with your expected values
    expectedWatcherConfig := sqldb.WatcherConfiguration{
        Location: "TestLocation",
        SecurityZone: "TestSecurityZone",
    }
    expectedMainConfig := sqldb.MainConfiguration{
		DBname:         "testDB",
		Hostname:       "testHost",
		IPAddress:      "192.168.1.1",
		ConfPath:       "/test/path",
		Sync:           true,
		UpdateInterval: 30,
		Hash:           "testhash",
		LastCheck:      1625076000,
		LastUpdate:     1625076000,
	}
	
    expectedPrometheusConfig := sqldb.PrometheusConfiguration{
        Url: "http://example.com",
        MetricName: "testMetric",
        Auth: true,
        Username: "user",
        Password: "pass",
        RetriesCount: 3,
    }
    expectedPrometheusLabel := sqldb.PrometheusLabelConfiguration{
		Opcode:             false,
		Authoritative:      false,
		Truncated:          true,
		Rcode:              true,
		RecursionDesired:   false,
		RecursionAvailable: false,
		AuthenticatedData:  false,
		CheckingDisabled:   false,
		PollingRate:        false,
		Recursion:          true,
	}	

	if !reflect.DeepEqual(WatcherConfig, expectedWatcherConfig) {
		t.Errorf("WatcherConfig does not match expected values. Got %+v, want %+v", WatcherConfig, expectedWatcherConfig)
	}
	if !reflect.DeepEqual(MainConfig, expectedMainConfig) {
		t.Errorf("MainConfig does not match expected values. Got %+v, want %+v", MainConfig, expectedMainConfig)
	}
	if !reflect.DeepEqual(PrometheusConfig, expectedPrometheusConfig) {
		t.Errorf("PrometheusConfig does not match expected values. Got %+v, want %+v", PrometheusConfig, expectedPrometheusConfig)
	}
	if !reflect.DeepEqual(PrometheusLabel, expectedPrometheusLabel) {
		t.Errorf("PrometheusLabel does not match expected values. Got %+v, want %+v", PrometheusLabel, expectedPrometheusLabel)
	}	
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
