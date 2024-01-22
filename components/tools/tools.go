package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)


func FileExists(filename string) bool {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return err == nil
}

func CheckPortAvailability(port int) bool {
    ln, err := net.Listen("tcp", ":" + strconv.Itoa(port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

func FormatUnixTime(unixTime int64) string {
    if unixTime == 0 {
        return "N/A"
    }
    t := time.Unix(unixTime, 0)
    return t.Format("2006-01-02 15:04:05") // You can change this format as needed
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
        // Write into log so wrong fetch IP
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
    // Write into log so wrong fetch IP
	return ""
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

func TruncateToken(token string) string {
    if len(token) < 10 {
        return "invalid_token"
    }
    return token[:10] + "..."
}

func ExtractToken(r *http.Request) string {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return "token_not_found"
    }
    return TruncateToken(cookie.Value)
}

func CalculateHash(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := sha256.New()

    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    
    return hex.EncodeToString(hash.Sum(nil)), nil
}
