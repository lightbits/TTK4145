package main

import (
    "fmt"
    "time"
    "./network"
)

const MASTER_UPDATE_INTERVAL  = 3 * time.Second
const CLIENT_TIMEOUT_INTERVAL = 5 * time.Second

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

    outgoing := make(chan network.MasterUpdate)
    incoming := make(chan network.ClientUpdate)
    go network.InitMaster(outgoing, incoming)

    client_timeout := make(chan Client)
    ticker := time.NewTicker(MASTER_UPDATE_INTERVAL)

    for {
        select {
        case update := <- incoming:
            sender_addr := update.Sender.String()

            if client, exists := Clients[sender_addr]; exists {

                fmt.Println("CLIENT", sender_addr, "said", update.Request)
                client.Timer.Reset(CLIENT_TIMEOUT_INTERVAL)

            } else {

                fmt.Println("CLIENT", sender_addr, "connected")
                new_client := Client{sender_addr, time.NewTimer(CLIENT_TIMEOUT_INTERVAL)}
                Clients[sender_addr] = new_client
                go ListenForClientTimeout(&new_client, client_timeout)

            }

        case <- ticker.C:
            outgoing <- network.MasterUpdate{"orders orders orders..."}
            fmt.Println("MASTER send update")

        case client := <- client_timeout:
            fmt.Println(client.Addr, "timed out")
        }
    }
}
