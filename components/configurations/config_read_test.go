package config

import (
	sqldb "HighFrequencyDNSChecker/components/db"
	"crypto/md5"
	"os"
	"testing"
	"time"
)

func TestLoadMainConfig(t *testing.T) {
    // Ensure proper YAML formatting with spaces for indentation
    yamlContent := `
General:
  db_name: "app.db"
  confCheckInterval: 1
  sync: true

Log:
  path: "log.json"
  minSeverity: "info"
  maxAge: 30
  maxSize: 10
  maxFiles: 10

Audit:
  path: "audit.json"
  minSeverity: "info"
  maxAge: 30
  maxSize: 10
  maxFiles: 10

WebServer:
  port: 443
  sslIsEnable: true
  sslCertPath: "cert.pem"
  sslKeyPath: "key.pem"
  sesionTimeout: 600
  initUsername: "admin"
  initPassword: "Samsung1"

Sync:
  isEnable: true
  members:
    - hostname: "127.0.0.1"
      port: 443
    - hostname: "10.10.10.10"
      port: 8081

Prometheus:
  url: "http://prometheus:8428/api/v1/write"
  metricName: "dns_resolve"
  auth: false
  username: "user"
  password: "password"
  retriesCount: 2
  bufer_size: 2

PrometheusLabels:
  opcode: false
  authoritative: false
  truncated: true
  rcode: true
  recursionDesired: false
  recursionAvailable: false
  authenticatedData: false
  checkingDisabled: false
  pollingRate: false
  recursion: true

Resolvers:
  path: "dns_servers.csv"
  pullTimeout: 2
  delimeter: ","
  extraDelimeter: "&"

Watcher:
  location: K2
  securityZone: PROD
`

    // Create a temporary YAML file
    tmpfile, err := os.CreateTemp("", "config.*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.WriteString(yamlContent); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    // Call LoadMainConfig with the temp file path
    config, err := LoadMainConfig(tmpfile.Name())
    if err != nil {
        t.Fatalf("LoadMainConfig() error = %v", err)
    }

    // Assertions
    if config.General.DBname != "app.db" {
        t.Errorf("Expected DBname 'app.db', got '%s'", config.General.DBname)
    }
    // ... Add more assertions as needed
}

func TestSetAdditionalInfoForResolvers(t *testing.T) {
    // Create a temporary file with known content using os.CreateTemp
    content := "known content for hashing"
    tmpfile, err := os.CreateTemp("", "resolver_config.*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name()) // clean up

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    // Set up the initial configuration
    conf := sqldb.ResolversConfiguration{
        Path: tmpfile.Name(),
    }

    // Set a known timestamp
    timestamp := time.Now().Unix()
    updatedConf, err := SetAdditionalInfoForResolvers(conf, timestamp)
    if err != nil {
        t.Fatalf("SetAdditionalInfoForResolvers() error = %v", err)
    }

    // Check if the LastCheck and LastUpdate fields are updated correctly
    if updatedConf.LastCheck != timestamp {
        t.Errorf("Expected LastCheck '%d', got '%d'", timestamp, updatedConf.LastCheck)
    }
    if updatedConf.LastUpdate != timestamp {
        t.Errorf("Expected LastUpdate '%d', got '%d'", timestamp, updatedConf.LastUpdate)
    }

    // Calculate the expected hash for the known content
    expectedHash, err := CalculateHash(tmpfile.Name(), md5.New)
    if err != nil {
        t.Fatal(err)
    }

    // Check if the Hash is calculated correctly
    if updatedConf.Hash != expectedHash {
        t.Errorf("Expected Hash '%s', got '%s'", expectedHash, updatedConf.Hash)
    }
}