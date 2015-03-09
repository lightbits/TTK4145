package main

import (
    "fmt"
    "time"
    "log"
    "net"
)

const MASTER_UPDATE_INTERVAL  = 3 * time.Second
const CLIENT_TIMEOUT_INTERVAL = 2 * time.Second

type Client struct {
    Address  *net.UDPAddr
    Timer    *time.Timer
    TimedOut bool
}

type OrderButton struct {
    Floor int32
    Type  int32
}

type IncomingClientStatus struct {
    SenderAddress *net.UDPAddr
    Data string
}

type MasterToClient struct {
    Data string
}

func ListenForClientTimeout(client *Client, timeout chan Client) {
    for {
        select {
        case <- client.Timer.C:
            timeout <- *client
        }
    }
}

func listenForClientStatus(conn *net.UDPConn, incoming chan IncomingClientStatus) {
    for {
        data := make([]byte, 1024)
        read_bytes, client_addr, err := conn.ReadFromUDP(data)
        if err == nil {
            incoming <- IncomingClientStatus{client_addr, string(data[:read_bytes])}
        } else {
            log.Println(err)
            return
        }
    }
}

func sendToClient(conn *net.UDPConn, client_addr *net.UDPAddr, packet MasterToClient) {
    _, err := conn.WriteToUDP([]byte(packet.Data), client_addr)
    if err != nil {
        log.Println(err)
    }
}

func main() {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    clients := make(map[string]Client)

    incoming := make(chan IncomingClientStatus)
    go listenForClientStatus(conn, incoming)

    client_timeout := make(chan Client)
    ticker := time.NewTicker(MASTER_UPDATE_INTERVAL)

    for {
        select {
        case update := <- incoming:

            map_key := update.SenderAddress.String()

            if client, exists := clients[map_key]; exists {
                client.Timer.Reset(CLIENT_TIMEOUT_INTERVAL)
            } else {
                fmt.Println("CLIENT", map_key, "connected")
                new_client := Client{update.SenderAddress,
                    time.NewTimer(CLIENT_TIMEOUT_INTERVAL),
                    false}
                clients[map_key] = new_client
                go ListenForClientTimeout(&new_client, client_timeout)
            }

            fmt.Println("CLIENT", map_key, "said", update.Data)

        case <- ticker.C:

            for _, client := range(clients) {
                var data MasterToClient
                data.Data = "Hey ho!"
                sendToClient(conn, client.Address, data)
            }
            fmt.Println("MASTER send update")

        case client := <- client_timeout:
            client.TimedOut = true
            fmt.Println(client.Address, "timed out")

        }
    }
}
