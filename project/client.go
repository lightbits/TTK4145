package main

import (
    "fmt"
    "time"
    "net"
    // "bytes"
    // "encoding/binary"
    "log"
)

const CLIENT_UPDATE_INTERVAL = 1 * time.Second

type Status struct {
    Data string
}

type OrderButton struct {
    Floor int32
    Type  int32
}

type MasterToClient struct {
    Data string
}

func broadcastStatusToMaster(conn *net.UDPConn, status Status) {
    remote, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.WriteToUDP([]byte(status.Data), remote)
    if err != nil {
        log.Fatal(err)
    }
}

func getUpdatesFromMaster(conn *net.UDPConn, incoming chan MasterToClient) {
    for {
        data := make([]byte, 1024)
        read_bytes, _, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }

        // TODO: Validate incoming packet
        // Check protocol etc

        // b := bytes.NewBuffer(data[:read_bytes])
        // r := MasterToClient{}
        // binary.Read(b, binary.BigEndian, &r)
        r := MasterToClient{string(data[:read_bytes])}

        incoming <- r
    }
}

func main() {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:54321")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()
    incoming := make(chan MasterToClient)
    go getUpdatesFromMaster(conn, incoming)
    ticker := time.NewTicker(CLIENT_UPDATE_INTERVAL)

    for {
        select {
        case <- ticker.C:
            fmt.Println("Client send update")
            broadcastStatusToMaster(conn, Status{"Hello"})

        case update := <- incoming:
            fmt.Println("Master said:", update.Data)
        }
    }
}
