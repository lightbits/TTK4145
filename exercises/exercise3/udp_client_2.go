package main

import (
    "net"
    "time"
    "log"
)

func main() {
    // Listen for data through this address:port
    local, err :=  net.ResolveUDPAddr("udp", "127.0.0.1:63303")
    if err != nil {
        log.Fatal(err)
    }

    // Send data to this address:port (broadcast address)
    remote, err := net.ResolveUDPAddr("udp", "255.255.255.255:30000")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialUDP("udp", local, remote)
    if err != nil {
        log.Fatal(err)
    }

    for {
        data := []byte("Testing 123")
        sent_bytes, err := conn.Write(data)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Sent", sent_bytes, "bytes")
        time.Sleep(1 * time.Second)
    }
}