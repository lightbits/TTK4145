package main

import (
    "fmt"
    "time"
    "log"
    "net"
    "os"
    "encoding/json"
)

const MASTER_UPDATE_INTERVAL  = 3 * time.Second
const CLIENT_TIMEOUT_INTERVAL = 2 * time.Second

type Client struct {
    Address         *net.UDPAddr
    Timer           *time.Timer
    TimedOut        bool
    LastPassedFloor int
    TargetFloor     int
}

type ButtonType int
const (
    ButtonUp ButtonType = iota
    ButtonDown
    ButtonOut
)

type OrderButton struct {
    Floor int
    Type  ButtonType
}

type Order struct {
    Button  OrderButton
    TakenBy string
}

type ClientStatus struct {
    SenderAddress   *net.UDPAddr
    LastPassedFloor int
    ClearedFloors   []int
    Commands        []OrderButton
}

type MasterUpdate struct {
    LitButtonLamps []OrderButton
    TargetFloor    int
}

func ListenForClientTimeout(client *Client, timeout chan *Client) {
    for {
        select {
        case <- client.Timer.C:
            timeout <- client
        }
    }
}

func listenToClient(conn *net.UDPConn, incoming chan ClientStatus) {
    for {
        bytes := make([]byte, 1024)
        read_bytes, client_addr, err := conn.ReadFromUDP(bytes)
        if err != nil {
            log.Println(err)
            continue
        }

        type Status struct {
            LastPassedFloor int
            ClearedFloors   []int
            Commands        []OrderButton
        }

        var status Status
        err = json.Unmarshal(bytes[:read_bytes], &status)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- ClientStatus{
            client_addr,
            status.LastPassedFloor,
            status.ClearedFloors,
            status.Commands}
    }
}

func sendToClient(conn *net.UDPConn, client_addr *net.UDPAddr, packet MasterUpdate) {
    bytes, err := json.Marshal(packet)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.WriteToUDP(bytes, client_addr)
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

    clients := make(map[string]*Client)
    orders  := make([]Order, 0)

    // Event channels
    incoming_update  := make(chan ClientStatus)
    client_timed_out := make(chan *Client)
    time_to_send     := time.NewTicker(MASTER_UPDATE_INTERVAL)
    time_to_display  := time.NewTicker(1 * time.Second)

    go listenToClient(conn, incoming_update)

    for {
        select {
        case status := <- incoming_update:

            id := status.SenderAddress.String()


            if _, exists := clients[id]; !exists {
                var c Client
                c.Address = status.SenderAddress
                c.Timer = time.NewTimer(CLIENT_TIMEOUT_INTERVAL)
                clients[id] = &c
                go ListenForClientTimeout(&c, client_timed_out)
            }
            clients[id].TimedOut = false
            clients[id].Timer.Reset(CLIENT_TIMEOUT_INTERVAL)
            clients[id].LastPassedFloor = status.LastPassedFloor

            for _, button := range(status.Commands) {
                exists := false
                for _, o := range(orders) {
                    exists = o.Button.Floor == button.Floor &&
                             o.Button.Type  == button.Type
                }
                if !exists {
                    orders = append(orders, Order{button, id})
                }
            }

            for _, floor := range(status.ClearedFloors) {
                for i, o := range(orders) {
                    if o.Button.Floor == floor {
                        orders = append(orders[:i], orders[i+1:]...)
                    }
                }
            }

        case <- time_to_send.C:

            for _, client := range(clients) {
                var data MasterUpdate
                data.LitButtonLamps = []OrderButton{
                    OrderButton{5, ButtonUp},
                    OrderButton{3, ButtonDown},
                    OrderButton{3, ButtonOut},
                }
                data.TargetFloor = 3
                sendToClient(conn, client.Address, data)
            }

        case client := <- client_timed_out:
            client.TimedOut = true

        case <- time_to_display.C:

            os.Stdout.Write([]byte("\033[2J"))
            os.Stdout.Write([]byte("\033[H"))

            fmt.Printf("Pending orders\n")
            fmt.Printf("Number\tFloor\tType\n")
            for n, o := range(orders) {
                s := "out"
                if o.Button.Type == ButtonUp {
                    s = "up"
                } else if o.Button.Type == ButtonDown {
                    s = "down"
                }
                fmt.Printf("%d\t%d\t%s\n", n, o.Button.Floor, s)
            }
            fmt.Printf("\nConnections\nIP Address\tTimed out\tLast floor\tOrders\n")
            for _, client := range(clients) {
                who := client.Address.String()
                timed_out := client.TimedOut
                last_floor := client.LastPassedFloor
                fmt.Printf("%s\t%v\t\t%d", who, timed_out, last_floor)
            }

        }
    }
}
