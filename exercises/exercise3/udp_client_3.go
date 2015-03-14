package main

import (
    "log"
    "net"
    "time"
)

func Listen(done chan bool) {
    // The address we wish to listen to
    local, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    // Create a socket
    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    // Read forever
    for {
        buffer := make([]byte, 1024)
        bytes_read, sender, err := conn.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Read", bytes_read, "from", sender, ":", string(buffer))
        time.Sleep(1 * time.Second)
    }

    done <- true
}

func Write(done chan bool) {
    // Server address
    remote, err := net.ResolveUDPAddr("udp", "129.241.187.255:20012")
    if err != nil {
        log.Fatal(err)
    }

    // Create a "connection" socket, use arbitrary local port
    conn, err := net.DialUDP("udp", nil, remote)
    if err != nil {
        log.Fatal(err)
    }

    // Send forever
    for {
        // Try to send something
        buffer := []byte("Hoopdoopawdop")
        bytes_sent, err := conn.Write(buffer)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Sent", bytes_sent)
        time.Sleep(1 * time.Second)
    }

    done <- true
}

func main() {
    done := make(chan bool)
    go Write(done)
    go Listen(done)

    <-done
    <-done
}
