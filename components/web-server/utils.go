package webserver

import (
    "net"
)

func CheckPortAvailability(port string) bool {
    ln, err := net.Listen("tcp", ":" + port)
    if err != nil {
        return false
    }
    ln.Close()
    return true
}
