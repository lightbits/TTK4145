package main

import (
    "net"
    "log"
)

type OrderType int32
const (
    OrderUp OrderType = iota
    OrderDown
    OrderOut
)

type LiftID struct {
    IP   uint32
}

type MasterOrder struct {
    Floor   int32
    Type    OrderType
    TakenBy LiftID
}

type MasterUpdate struct {
    ActiveOrders []MasterOrder
}

type ClientRequest struct {
    Floor int32
    Type  OrderType
    Done  bool
}

type ClientUpdate struct {
    Requests []ClientRequest
}

func ListenForMasterUpdates(incoming chan MasterUpdate) {

}

func SendMasterUpdates(outgoing chan MasterUpdate, conn *net.UDPConn) {

}

func Master() {
    Connections  []net.UDPConn
    ActiveOrders []MasterOrder

    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
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

func main() {

}
