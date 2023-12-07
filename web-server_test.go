package main

import (
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "strings"
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

    // Start the server
    go webserver.Webserver()

    // Create a client with a cookie jar to handle cookies
    jar, err := cookiejar.New(nil)
    if err != nil {
        t.Fatalf("Failed to create cookie jar: %v", err)
    }
    client := &http.Client{Jar: jar}

    // Login
    loginURL := "http://localhost:" + port + "/login"
    loginData := url.Values{}
    loginData.Set("username", "user")
    loginData.Set("password", "Samsung1")
    req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData.Encode()))
    if err != nil {
        t.Fatalf("Failed to create login request: %v", err)
    }
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("Failed to login: %v", err)
    }
    defer resp.Body.Close()

    // Check if login was successful and token was received
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("Login failed, status: %s", resp.Status)
    }

    // Now make a request to your endpoint with the token
    testURL := "http://localhost:" + port // replace with your actual endpoint URL
    resp, err = client.Get(testURL)
    if err != nil {
        t.Fatalf("Failed to make request with token: %v", err)
    }
    defer resp.Body.Close()

    // Check if the request was successful
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK (200); got %v", resp.Status)
    }
}
