package main

import (
    "log"
    "net"
    // "time"
)

func ListenForConnections(done chan bool) {
    local, err := net.ResolveTCPAddr("tcp", ":12345")
    if err != nil {
        log.Fatal(err)
    }

    listener, err := net.ListenTCP("tcp", local)
    if err != nil {
        log.Fatal(err)
    }

    conn, err := listener.AcceptTCP()
    if err != nil {
        log.Fatal(err)
    }

    log.Println(conn.RemoteAddr(), "connected to arbeidsplass 12!")
    conn.Close()
    done <- true
}

func main() {
    remote, err := net.ResolveTCPAddr("tcp", "129.241.187.136:33546")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialTCP("tcp", nil, remote)
    if err != nil {
        log.Fatal(err)
    }

    buffer := make([]byte, 1024)
    bytes_read, err := conn.Read(buffer)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Read", bytes_read, "bytes:", string(buffer))

    done := make(chan bool)
    go ListenForConnections(done)

    bytes_sent, err := conn.Write([]byte("Connect to: 129.241.187.144:12345"))
    conn.Write([]byte{0})
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Sent", bytes_sent, "bytes")

    <- done
    err = conn.Close()
    if err != nil {
        log.Fatal(err)
    }
}