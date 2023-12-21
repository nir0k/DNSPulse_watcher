package config

import (
	"fmt"
	"hash"
	"io"
	"net"
	"os"
	"regexp"
)


func CalculateHash(filePath string, hashFunc func() hash.Hash) (string, error) {
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


func ValidateFilePath(path string) bool {
	validPathRegex := regexp.MustCompile("^[a-zA-Z0-9- _/.]+$")
	return validPathRegex.MatchString(path)
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
