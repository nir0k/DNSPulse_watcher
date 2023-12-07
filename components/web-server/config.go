package webserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
    "github.com/nir0k/HighFrequencyDNSChecker/components/log"
)


var (
    currentPort string
    server      *http.Server
    lock        sync.Mutex
)


func GetEnvVariable(key string) (string, error) {
    envContent, err := os.ReadFile(".env")
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(envContent), "\n")
    for _, line := range lines {
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 && parts[0] == key {
            return parts[1], nil
        }
    }

    return "", fmt.Errorf("variable %s not found in .env file", key)
}


func WatchForPortChanges(done *chan bool) {
    for {
        time.Sleep(1 * time.Minute)
        port, err := GetEnvVariable("WATCHER_WEB_PORT")
        if err != nil {
            fmt.Println("Error reading port:", err)
            continue
        }

        lock.Lock()
        if port != currentPort {
            currentPort = port
            if server != nil {
                fmt.Println("Changing port to:", port)
                close(*done)

                newDone := make(chan bool)
                *done = newDone

                go StartServer(port, newDone)
            }
        }
        lock.Unlock()
    }
}


func StartServer(port string, done chan bool) {
    if !CheckPortAvailability(port) {
        log.AppLog.Error("Port is already in use. Cannot start the web server. Port:", port)
        return
    }

    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/", authMiddleware(homeHandler))
    http.HandleFunc("/csv", authMiddleware(csvHandler))
    http.HandleFunc("/csv/upload", authMiddleware(uploadCSVHandler))
    http.HandleFunc("/csv/download", authMiddleware(downloadCSVHandler))
    http.HandleFunc("/csv/delete", deleteCsvRowHandler)
    http.HandleFunc("/csv/edit", editCsvRowHandler)
    http.HandleFunc("/env", authMiddleware(envHandler))
	http.HandleFunc("/logout", authMiddleware(logoutHandler))

    server = &http.Server{Addr: ":" + port, Handler: nil}
    go func() {
        <-done
        if err := server.Shutdown(context.Background()); err != nil {
            fmt.Println("Server Shutdown:", err)
        }
    }()

    fmt.Println("Server starting on port", port)
    err := server.ListenAndServe()
    if err != http.ErrServerClosed {
        fmt.Println("Server failed:", err)
    }
}
