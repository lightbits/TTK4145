package main

import (
    "fmt"
    "time"
    "encoding/binary"
    "bytes"
    "./network"
)

const MASTER_UPDATE_INTERVAL  = 3 * time.Second
const CLIENT_TIMEOUT_INTERVAL = 2 * time.Second

type Client struct {
    ID    string // TODO: LiftID
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

type LiftID uint32

type Order struct {
    Floor   int32
    TakenBy LiftID
}

func (o Order) ByteSerialize() *bytes.Buffer {
    b := &bytes.Buffer{}
    err := binary.Write(b, binary.BigEndian, o)
    if err != nil {
        fmt.Println(err)
    }
    return b
}

func main() {
    // TODO: Make key LiftID
    clients := make(map[string]Client)

    var active_orders [2]Order
    active_orders[0] = Order{
        Floor: 5,
        TakenBy: 0xdeadbeef}

    active_orders[1] = Order{
        Floor: 2,
        TakenBy: 0xabad1dea}

    outgoing := make(chan network.MasterUpdate)
    incoming := make(chan network.ClientUpdate)
    go network.InitMaster(outgoing, incoming)

    client_timeout := make(chan Client)
    ticker := time.NewTicker(MASTER_UPDATE_INTERVAL)

    for {
        select {
        case update := <- incoming:
            client_id := update.Sender.String()

            if client, exists := clients[client_id]; exists {

                fmt.Println("CLIENT", client_id, "said", update.Request)
                client.Timer.Reset(CLIENT_TIMEOUT_INTERVAL)

            } else {

                fmt.Println("CLIENT", client_id, "connected")
                new_client := Client{client_id, time.NewTimer(CLIENT_TIMEOUT_INTERVAL)}
                clients[client_id] = new_client
                go ListenForClientTimeout(&new_client, client_timeout)

            }

        case <- ticker.C:

            // Write to byte array
            b1 := active_orders[0].ByteSerialize().Bytes()
            b2 := active_orders[1].ByteSerialize().Bytes()
            fmt.Printf("%x\n", append(b1[:], b2[:]...))

            outgoing <- network.MasterUpdate{"orders orders orders..."}
            fmt.Println("MASTER send update")

        case client := <- client_timeout:
            fmt.Println(client.ID, "timed out")
        }
    }
}
