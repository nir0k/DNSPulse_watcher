package main

import (
	"github.com/nir0k/HighFrequencyDNSChecker/components/watcher"
	"os"
	"strings"
	"testing"
	"time"
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
LOG_MAX_AGE=30
LOG_MAX_SIZE=10
LOG_MAX_FILES=10
WATCHER_WEB_AUTH_LOG_FILE=high_frequency_dns_mon_auth.log
WATCHER_WEB_AUTH_LOG_LEVEL=info
WATCHER_WEB_AUTH_LOG_MAX_AGE=30
WATCHER_WEB_AUTH_LOG_MAX_SIZE=10
WATCHER_WEB_AUTH_LOG_MAX_FILES=10
CONF_CHECK_INTERVAL=1
BUFFER_SIZE=2
WATCHER_LOCATION=K
WATCHER_SECURITYZONE=PROD




`
	tempEnvFile := "temp.env"
	if err := os.WriteFile(tempEnvFile, []byte(envContent), 0644); err != nil {
        t.Fatalf("Failed to create temporary .env file: %v", err)
    }

	success := watcher.ReadConfig(tempEnvFile)

	if !success {
		t.Errorf("readConfig failed")
	}

	if err := os.Remove(tempEnvFile); err != nil {
		t.Fatalf("Failed to delete temporary .env file: %v", err)
	}
}


func TestDnsResolveServiceMode(t *testing.T) {
    resolver := watcher.Resolver{
        Server: "test_server",
        Server_ip: "1.2.3.4",
        Service_mode: true,
    }

    performed := watcher.DnsResolve(resolver)

    if performed {
        t.Errorf("dnsResolve should not perform resolution when service_mode is true")
    }
}

func TestCheckConfigWithEnvChange(t *testing.T) {
    cwd, _ := os.Getwd()
    t.Log("Current working directory:", cwd)

    watcher.Config.Conf_path = cwd + "/.env"
    t.Log("Config path:", watcher.Config.Conf_path)

    originalEnv, err := os.ReadFile(watcher.Config.Conf_path)
    if err != nil {
        t.Fatalf("Failed to read .env file: %v", err)
    }

    watcher.Setup()

    modifiedEnv := modifyEnvParameter(string(originalEnv), "LOG_LEVEL", "info")

    defer func() {
        if err := os.WriteFile(watcher.Config.Conf_path, originalEnv, 0644); err != nil {
            t.Fatalf("Failed to restore .env file: %v", err)
        }
    }()

    time.Sleep(3 * time.Second)

    if err := os.WriteFile(watcher.Config.Conf_path, []byte(modifiedEnv), 0644); err != nil {
        t.Fatalf("Failed to modify .env file: %v", err)
    }

    time.Sleep(3 * time.Second)

    confCompare, resolversState := watcher.CheckConfig()

    if confCompare {
        t.Errorf("Expected configuration changes, but config was not marked as changed")
    }
    if !resolversState {
        t.Errorf("Expected DNS resolvers to be unchanged, but they were marked as changed")
    }
}

func TestCheckConfigWithCSVChange(t *testing.T) {
    cwd, _ := os.Getwd()
    t.Log("Current working directory:", cwd)

    watcher.Config.Conf_path = cwd + "/.env"
    watcher.ReadConfig(watcher.Config.Conf_path)

    csvPath := cwd + "/" + watcher.Dns_param.Dns_servers_path
    t.Log("CSV path:", csvPath)

    originalCSV, err := os.ReadFile(csvPath)
    if err != nil {
        t.Fatalf("Failed to read DNS servers CSV file: %v", err)
    }

    modifiedCSV := string(originalCSV) + "\nnew_server,1.2.3.4,false,newdomain.com,prefix,location,site,zone,udp,zone1,10"

    defer func() {
        if err := os.WriteFile(csvPath, originalCSV, 0644); err != nil {
            t.Fatalf("Failed to restore DNS servers CSV file: %v", err)
        }
    }()

    if err := os.WriteFile(csvPath, []byte(modifiedCSV), 0644); err != nil {
        t.Fatalf("Failed to modify DNS servers CSV file: %v", err)
    }

    time.Sleep(3 * time.Second)

    confCompare, resolversState := watcher.CheckConfig()

    if !confCompare {
        t.Errorf("Expected configuration changes, but config was not marked as changed")
    }
    if resolversState {
        t.Errorf("Expected DNS resolvers to be marked as changed, but they were not")
    }
}


func modifyEnvParameter(envContent, parameter, newValue string) string {
    lines := strings.Split(envContent, "\n")
    for i, line := range lines {
        if strings.HasPrefix(line, parameter+"=") {
            lines[i] = parameter + "=" + newValue
            break
        }
    }
    return strings.Join(lines, "\n")
}
