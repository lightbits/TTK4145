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
    Address *net.UDPAddr
    Timer   *time.Timer
}

type IncomingClientStatus struct {
    SenderAddress *net.UDPAddr
    Data string
}

type ClientTodo struct {
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

func listenForIncomingClientStatuses(conn *net.UDPConn, incoming chan IncomingClientStatus) {
    for {
        data := make([]byte, 1024)
        read_bytes, client_addr, err := conn.ReadFromUDP(data)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- IncomingClientStatus{client_addr, string(data[:read_bytes])}
    }
}

func sendTodosToClient(conn *net.UDPConn, todo ClientTodo, destination *net.UDPAddr) {
    conn.WriteToUDP([]byte(todo.Data), destination)
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

    // var active_orders [2]Order
    // active_orders[0] = Order{
    //     Floor: 5,
    //     TakenBy: 0xdeadbeef}

    // active_orders[1] = Order{
    //     Floor: 2,
    //     TakenBy: 0xabad1dea}

    incoming := make(chan IncomingClientStatus)
    go listenForIncomingClientStatuses(conn, incoming)

    client_timeout := make(chan Client)
    ticker := time.NewTicker(MASTER_UPDATE_INTERVAL)

    for {
        select {
        case update := <- incoming:
            map_key := update.SenderAddress.String()

            if client, exists := clients[map_key]; exists {

                fmt.Println("CLIENT", map_key, "said", update.Data)
                client.Timer.Reset(CLIENT_TIMEOUT_INTERVAL)

            } else {

                fmt.Println("CLIENT", map_key, "connected")
                new_client := Client{update.SenderAddress, time.NewTimer(CLIENT_TIMEOUT_INTERVAL)}
                clients[map_key] = new_client
                go ListenForClientTimeout(&new_client, client_timeout)

            }

        case <- ticker.C:

            // Write to byte array
            // b1 := active_orders[0].ByteSerialize().Bytes()
            // b2 := active_orders[1].ByteSerialize().Bytes()
            // fmt.Printf("%x\n", append(b1[:], b2[:]...))

            for _, client := range(clients) {
                sendTodosToClient(conn, ClientTodo{"lamps lamps"}, client.Address)
            }
            fmt.Println("MASTER send update")

        case client := <- client_timeout:
            fmt.Println(client.Address, "timed out")
        }
    }
}
