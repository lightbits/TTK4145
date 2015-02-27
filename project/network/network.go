package network

import (
    "net"
    "log"
)

type MasterUpdate struct {
    ActiveOrders string
}

type ClientUpdate struct {
    Sender  *net.UDPAddr
    Request string
}

func sendClientUpdates(outgoing chan ClientUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing
        remote, err := net.ResolveUDPAddr("udp", "255.255.255.255:20012")
        if err != nil {
            log.Fatal(err)
        }
        conn.WriteToUDP([]byte(update.Request), remote)
    }
}

func listenForMasterUpdates(incoming chan MasterUpdate, conn *net.UDPConn) {
    for {
        data := make([]byte, 1024)
        read_bytes, _, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }

        // TODO: Validate incoming packet
        // Check protocol etc

        incoming <- MasterUpdate{string(data[:read_bytes])}
    }
}

func listenForClientUpdates(incoming chan ClientUpdate, conn *net.UDPConn) {
    for {
        data := make([]byte, 1024)
        read_bytes, sender_addr, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- ClientUpdate{sender_addr, string(data[:read_bytes])}
    }
}

func sendMasterUpdates(outgoing chan MasterUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing

        // TODO: Should we send to each connection seperately instead?
        // For now: Broadcast
        remote, err := net.ResolveUDPAddr("udp", "255.255.255.255:54321")
        if err != nil {
            log.Fatal(err)
        }
        conn.WriteToUDP([]byte(update.ActiveOrders), remote)
    }
}

func InitClient(outgoing chan ClientUpdate, incoming chan MasterUpdate) {
    local, err := net.ResolveUDPAddr("udp", ":54321")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    go sendClientUpdates(outgoing, conn)
    listenForMasterUpdates(incoming, conn)
}

func InitMaster(outgoing chan MasterUpdate, incoming chan ClientUpdate) {
    local, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    go sendMasterUpdates(outgoing, conn)
    listenForClientUpdates(incoming, conn)
}
