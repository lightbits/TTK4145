package main

import (
    "net"
    "time"
    "log"
    // "fmt"
)

func ListenThing(conn *net.UDPConn, read_channel chan int) {
    for {
        data := make([]byte, 1024)
        log.Println("Attempt read")
        read_bytes, _, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        read_channel <- read_bytes
        // read_channel <- fmt.Sprintf("Read %d bytes from %q", read_bytes, sender)
        // log.Println("Read", read_bytes, "bytes from", sender)
    }
}

func WriteThing(conn *net.UDPConn) {
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

func main() {
    // Listen for data through this address:port
    local, err :=  net.ResolveUDPAddr("udp", "127.0.0.1:63303")
    if err != nil {
        log.Fatal(err)
    }

    // Send data to this address:port (broadcast address)
    remote, err := net.ResolveUDPAddr("udp", "127.0.0.1:30000")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialUDP("udp", local, remote)
    if err != nil {
        log.Fatal(err)
    }

    read_channel := make(chan int)

    go ListenThing(conn, read_channel)
    go WriteThing(conn)

    for {
        log.Println(<- read_channel)
    }
}