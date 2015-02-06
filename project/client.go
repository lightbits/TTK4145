package main

import (
    "log"
    "net"
    "time"
)

func BroadcastPing(done chan bool) {
    remote, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialUDP("udp", nil, remote)
    if err != nil {
        log.Fatal(err)
    }

    for {
        buffer := []byte("pI am pinging you dawg!")
        bytes_sent, err := conn.Write(buffer)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Sent", bytes_sent)

        readbuffer := make([]byte, 1024)
        _, err = conn.Read(readbuffer)
        if err != nil {
            log.Fatal(err)
        }
        log.Println(string(readbuffer))
        time.Sleep(1 * time.Second)
    }

    done <- true
}

func main() {
    done := make(chan bool)
    go BroadcastPing(done)
    
    <-done
}