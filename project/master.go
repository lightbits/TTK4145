package main

import (
    "net"
    "log"
    "fmt"
    "time"
)

const MASTER_UPDATE_INTERVAL  = 3 * time.Second
const CLIENT_TIMEOUT_INTERVAL = 5 * time.Second

type MasterUpdate struct {
    ActiveOrders string
}

type ClientUpdate struct {
    Sender  *net.UDPAddr
    Request string
}

func ListenForClientUpdates(incoming chan ClientUpdate, conn *net.UDPConn) {
    for {
        data := make([]byte, 1024)
        read_bytes, sender_addr, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- ClientUpdate{sender_addr, string(data[:read_bytes])}
    }
}

func SendMasterUpdates(outgoing chan MasterUpdate, conn *net.UDPConn) {
    for {
        update := <- outgoing

        // TODO: Should we send to each connection seperately instead?
        // For now: Broadcast
        remote, err := net.ResolveUDPAddr("udp", "255.255.255.255.54321")
        if err != nil {
            log.Fatal(err)
        }
        conn.WriteToUDP([]byte(update.ActiveOrders), remote)
    }
}

type Client struct {
    Addr   string
    Timer *time.Timer
}

func ListenForClientTimeout(client *Client, timeout chan Client) {
    for {
        select {
        case <- client.Timer.C:
            timeout <- *client
        }
    }
}

func main() {
    Clients := make(map[string]Client)

    local, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    outgoing := make(chan MasterUpdate)
    incoming := make(chan ClientUpdate)
    go ListenForClientUpdates(incoming, conn)
    go SendMasterUpdates(outgoing, conn)

    client_timeout := make(chan Client)
    ticker := time.NewTicker(MASTER_UPDATE_INTERVAL)

    for {
        select {
        case update := <- incoming:
            sender_addr := update.Sender.String()
            fmt.Println("CLIENT", sender_addr, "said", update.Request)

            if client, exists := Clients[sender_addr]; exists {

                fmt.Println("CLIENT", sender_addr, "pinged us. Resetting timer.")
                client.Timer.Reset(CLIENT_TIMEOUT_INTERVAL)

            } else {

                fmt.Println("CLIENT", sender_addr, "connected")
                new_client := Client{sender_addr, time.NewTimer(CLIENT_TIMEOUT_INTERVAL)}
                Clients[sender_addr] = new_client
                go ListenForClientTimeout(&new_client, client_timeout)

            }

        case <- ticker.C:
            fmt.Println("MASTER send update")

        case client := <- client_timeout:
            fmt.Println(client.Addr, "timed out")
        }
    }
}
