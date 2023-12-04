package webserver

import (
	"net/http"
	"os"
	"strings"
	"github.com/nir0k/HighFrequencyDNSChecker/components/log"
	"github.com/sirupsen/logrus"
)


func checkCredentials(username, password string) bool {
    envContent, err := os.ReadFile(".env")
    if err != nil {
        return false
    }

    lines := strings.Split(string(envContent), "\n")
    var envUsername, envPassword string
    for _, line := range lines {
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            if parts[0] == "WATCHER_WEB_USER" {
                envUsername = parts[1]
            } else if parts[0] == "WATCHER_WEB_PASSWORD" {
                envPassword = parts[1]
            }
        }
    }
    return username == envUsername && password == envPassword
}


func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth()
        clientIP := getClientIP(r)

        if !ok || !checkCredentials(username, password) {
            log.AuthLog.WithFields(logrus.Fields{
                "username": username,
                "ip":       clientIP,
                "status":   "failed",
            }).Warn("Authentication attempt failed")
            w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        log.AuthLog.WithFields(logrus.Fields{
            "username": username,
            "ip":       clientIP,
            "status":   "successful",
        }).Info("Authentication successful")
        handler(w, r)
    }
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("WWW-Authenticate", `Basic realm="loggedout"`)
    http.Error(w, "Logged out", http.StatusUnauthorized)
}


func getClientIP(r *http.Request) string {
    ip := r.Header.Get("X-REAL-IP")
    if ip == "" {
        ip = r.Header.Get("X-FORWARDED-FOR")
    }
    if ip == "" {
        ip = r.RemoteAddr
    }
    return ip
}