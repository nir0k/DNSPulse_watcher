package webserver

import (
	"net"
	"net/http"
	"time"
)

func CheckPortAvailability(port string) bool {
    ln, err := net.Listen("tcp", ":" + port)
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

func GetClientIP(r *http.Request) string {
    ip := r.Header.Get("X-REAL-IP")
    if ip == "" {
        ip = r.Header.Get("X-FORWARDED-FOR")
    }
    if ip == "" {
        ip = r.RemoteAddr
    }
    return ip
}

func ExtractToken(r *http.Request) string {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return "token_not_found"
    }
    return TruncateToken(cookie.Value) // Use truncateToken to partially hide the token
}


func TruncateToken(token string) string {
    if len(token) < 10 {
        return "invalid_token"
    }
    return token[:10] + "..."
}

func FormatUnixTime(unixTime int64) string {
    if unixTime == 0 {
        return "N/A"
    }
    t := time.Unix(unixTime, 0)
    return t.Format("2006-01-02 15:04:05") // You can change this format as needed
}