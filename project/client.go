package main

import (
    "net"
    "log"
    "fmt"
    "time"
)

type MasterUpdate struct {
    ActiveOrders string
}

type ClientUpdate struct {
    Request string
}

func SendClientUpdates(outgoing chan ClientUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing
        remote, err := net.ResolveUDPAddr("udp", "255.255.255.255:20012")
        if err != nil {
            log.Fatal(err)
        }
        conn.WriteToUDP([]byte(update.Request), remote)
    }
}

func main() {
    local, err := net.ResolveUDPAddr("udp", ":54321")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    outgoing := make(chan ClientUpdate)
    go SendClientUpdates(outgoing, conn)

    const CLIENT_UPDATE_INTERVAL = 1 * time.Second
    ticker := time.NewTicker(CLIENT_UPDATE_INTERVAL)

    for {
        select {
        case <- ticker.C:
            fmt.Println("Client send update")
            outgoing <- ClientUpdate{"Hello master!"}
        }
    }
}
