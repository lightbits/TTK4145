package network

import (
    "net"
    "log"
)

type ClientRequest struct {
    Data string
}

type MasterResponse struct {
    Sender
    Data string
}

type Client struct {
    ID string
}

func listenForMaster(socket *net.UDPConn, incoming chan MasterResponse) {
    for {
        data := make([]byte, 1024)
        read_bytes, _, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- MasterResponse{string(data[:read_bytes])}
    }
}

func ClientWorker(incoming chan MasterResponse, outgoing chan ClientRequest) {
    client, err := net.ResolveUDPAddr("udp", "127.0.0.1:54321")
    if err != nil {
        log.Fatal(err)
    }

    socket, err := net.ListenUDP("udp", client)
    if err != nil {
        log.Fatal(err)
    }

    master, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    if err != nil {
        log.Fatal(err)
    }

    go listenForMaster(incoming)

    for {
        r := <- outgoing
        _, err = conn.WriteToUDP([]byte(status.Data), master)
        if err != nil {
            log.Fatal(err)
        }
    }

    conn.Close()
}

func listenForClient(incoming chan ClientRequest) {

}

func MasterWorker(incoming chan ClientRequest, outgoing chan MasterResponse) {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    go listenForMaster(incoming)

    for {
        r := <- outgoing
        _, err = conn.WriteToUDP([]byte(status.Data), master)
        if err != nil {
            log.Fatal(err)
        }
    }

    conn.Close()
}
