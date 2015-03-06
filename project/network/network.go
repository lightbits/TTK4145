package network

import (
    "net"
    "log"
)

type OrderButton struct {
    Type  int32
    Floor int32
}

type MasterToClientUpdate struct {
    ActiveOrders string
    // Lamps       []OrderButton
    // TargetFloor int
}

type ClientToMasterUpdate struct {
    Request string
    // LiftCommands  []OrderButton
    // ClearedFloors []int
}

func broadcastClientUpdateToMaster(outgoing chan ClientToMasterUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing
        remote, err := net.ResolveUDPAddr("udp", "255.255.255.255:20012")
        if err != nil {
            log.Fatal(err)
        }
        conn.WriteToUDP([]byte(update.Request), remote)
    }
}

func listenForUpdatesFromMaster(incoming chan MasterToClientUpdate, conn *net.UDPConn) {

    for {
        data := make([]byte, 1024)
        read_bytes, _, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }

        // TODO: Validate incoming packet
        // Check protocol etc

        incoming <- MasterToClientUpdate{string(data[:read_bytes])}
    }
}

func listenForUpdatesFromClient(incoming chan ClientToMasterUpdate, conn *net.UDPConn) {
    for {
        data := make([]byte, 1024)
        read_bytes, sender_addr, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- ClientToMasterUpdate{sender_addr, string(data[:read_bytes])}
    }
}

func sendUpdatesToClient(outgoing chan MasterToClientUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing
        conn.WriteToUDP([]byte(update.ActiveOrders), update.Destination)
    }
}

func InitClient(outgoing chan ClientToMasterUpdate, incoming chan MasterToClientUpdate) {
    local, err := net.ResolveUDPAddr("udp", ":54321")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    go sendClientUpdatesToMaster(outgoing, conn)
    listenForUpdatesFromMaster(incoming, conn)
}

func InitMaster(outgoing chan MasterToClientUpdate, incoming chan ClientToMasterUpdate) {
    local, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    go sendUpdateToClient(outgoing, conn)
    listenForUpdatesFromClient(incoming, conn)
}
