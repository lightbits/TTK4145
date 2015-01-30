package main

import (
    "log"
    "net"
)

func main() {
    // Don't care about which IP address we are assigned
    // Only that we use port 30000
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:30000")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    // The connection conn can now be used to read and 
    // write messages.

    data := make([]byte, 1024) // Let's hope this is enough!

    for {
        read_bytes, sender, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }

        // Send a reply
        log.Println("Read", read_bytes, "bytes from", sender)
        conn.WriteToUDP([]byte("Server says: Gotcha!"), sender)
    }
}