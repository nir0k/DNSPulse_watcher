package main


import (
    "net/http"
    "testing"
	"github.com/nir0k/HighFrequencyDNSChecker/components/web-server"
    "github.com/nir0k/HighFrequencyDNSChecker/components/watcher"
)

func TestServerAccessibility(t *testing.T) {
    watcher.ReadConfig(".env")
    port, err := webserver.GetEnvVariable("WATCHER_WEB_PORT")
    if err != nil {
        t.Fatalf("Could not get port from .env file: %v", err)
    }
    username, err := webserver.GetEnvVariable("WATCHER_WEB_USER")
    if err != nil {
        t.Fatalf("Could not get username from .env file: %v", err)
    }
    password, err := webserver.GetEnvVariable("WATCHER_WEB_PASSWORD")
    if err != nil {
        t.Fatalf("Could not get password from .env file: %v", err)
    }

    go webserver.Webserver()

    resp, err := http.Get("http://localhost:" + port)
    if err != nil {
        t.Fatalf("Could not connect to server: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusUnauthorized {
        t.Errorf("Expected status Unauthorized (401); got %v", resp.Status)
    }

    client := &http.Client{}
    req, err := http.NewRequest("GET", "http://localhost:"+port, nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.SetBasicAuth(username, password)

    resp, err = client.Do(req)
    if err != nil {
        t.Fatalf("Failed to make authenticated request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK (200); got %v", resp.Status)
    }
}
