package main

import (
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {
	cwd, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current working directory: %v", err)
    }
    t.Logf("Current working directory: %s", cwd)
	envContent := `DNS_RESOLVERPATH=dns_servers.csv
DNS_TIMEOUT=1
DELIMETER=,
DELIMETER_FOR_ADDITIONAL_PARAM=&
PROM_URL=http://prometheus:8428/api/v1/write
PROM_METRIC=dns_resolve
PROM_AUTH=false
PROM_USER=testuser
PROM_PASS=testpass
PROM_RETRIES=2
OPCODES=false
AUTHORITATIVE=false
TRUNCATED=true
RCODE=true
RECURSION_DESIRED=false
RECURSION_AVAILABLE=false
AUTHENTICATE_DATA=false
CHECKING_DISABLED=false
POLLING_RATE=false
RECURSION=true
LOG_FILE=test.log
LOG_LEVEL=debug
CONF_CHECK_INTERVAL=1
BUFFER_SIZE=2
WATCHER_LOCATION=K
WATCHER_SECURITYZONE=PROD`
	tempEnvFile := "temp.env"
	if err := os.WriteFile(tempEnvFile, []byte(envContent), 0644); err != nil {
        t.Fatalf("Failed to create temporary .env file: %v", err)
    }

	success := readConfig(tempEnvFile)

	if !success {
		t.Errorf("readConfig failed")
	}

	if err := os.Remove(tempEnvFile); err != nil {
		t.Fatalf("Failed to delete temporary .env file: %v", err)
	}
}
