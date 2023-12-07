package watcher

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"sync"
    "github.com/nir0k/HighFrequencyDNSChecker/components/log"
)


var Mu sync.Mutex


func calculateHash(filePath string, hashFunc func() hash.Hash) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := hashFunc()
    _, err = io.Copy(hash, file)
    if err != nil {
        return "", err
    }

    hashSum := hash.Sum(nil)
    return fmt.Sprintf("%x", hashSum), nil
}


func compareFileHash(path string, curent_hash string) (bool, error) {
    
    new_hash, err := calculateHash(path, md5.New)
    if err != nil {
        log.AppLog.Error("Error: calculating MD5 hash to file '", path, "'. error:", err)
        return false, err
    }
    if curent_hash == new_hash {
        return true, nil
    }
    return false, nil
}


func isValidURL(inputURL string) bool {
    _, err := url.ParseRequestURI(inputURL)
    return err == nil
}


func isAlphaNumericWithDashOrUnderscore(input string) bool {
    validRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
    return validRegex.MatchString(input)
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no local IP address found")
}