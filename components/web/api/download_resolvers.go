package api

import (
	// "fmt"
	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)


func DownloadConfigurationHandler(w http.ResponseWriter, r *http.Request) {
    var (
		err error
		conf sqldb.ResolversConfiguration
	)

	conf, err = sqldb.GetResolverConfig(sqldb.AppDB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

    file, err := os.Open(conf.Path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    defer file.Close()

    // Extract filename for Content-Disposition header
    _, fileName := path.Split(conf.Path)

    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename="+fileName)

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.AuthLog.WithFields(logrus.Fields{
        "action": "csv_download",
        "file":   fileName,
        "ip":     getClientIP(r),
        "token":  extractToken(r),
    }).Info("CSV file downloaded")
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

func extractToken(r *http.Request) string {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return "token_not_found"
    }
    return truncateToken(cookie.Value) // Use truncateToken to partially hide the token
}


func truncateToken(token string) string {
    if len(token) < 10 {
        return "invalid_token"
    }
    return token[:10] + "..."
}
