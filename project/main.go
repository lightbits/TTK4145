package main

import (
    "time"
    "net"
    "log"
    "flag"
    "encoding/json"
    "./network"
    "./fakedriver"
)

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type Queue []Order

type ClientData struct {
    LastPassedFloor int
    Requests Queue
}

type MasterData struct {
    Orders Queue
}

type Channels struct {
    button_pressed  chan driver.OrderButton
    floor_reached   chan int
    stop_button     chan bool
    obstruction     chan bool
    incoming        chan network.IncomingPacket
    outgoing        chan network.OutgoingPacket
    outgoing_all    chan network.OutgoingPacket
}

func (q Queue) RemoveOrdersAtFloor(f int) {
    // TODO: Implement
    // for _, o := range(q.Orders) {
    //     if o.Floor == f {

    //     }
    // }
}

func DecodeMasterPacket(b []byte) MasterData {
    var result MasterData
    err := json.Unmarshal(b, &result)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func DecodeClientPacket(b []byte) ClientData {
    var result ClientData
    err := json.Unmarshal(b, &result)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func EncodeMasterData(m MasterData) []byte {
    result, err := json.Marshal(m)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func EncodeClientData(c ClientData) []byte {
    result, err := json.Marshal(c)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func WaitForBackup(c Channels) {
    log.Println("Waiting for backup")
    for {
        select {
        case packet := <- c.incoming:
            sender := packet.Sender
            Master(c, sender)
            return
        }
    }
}

type ClientType struct {
    ID              network.ID
    Timer          *time.Timer
    LastPassedFloor int
    TargetFloor     int
}

func ListenForClientTimeout(client *ClientType, timeout chan *ClientType) {
    for {
        select {
        case <- client.Timer.C:
            timeout <- client
        }
    }
}

func Master(c Channels, backup network.ID) {
    log.Println("Starting master with backup", backup)
}

func Backup(c Channels) {

}

func WaitForMaster(c Channels, q Queue) {
    log.Println("Waiting for master")
    time_to_ping := time.NewTicker(5*time.Second)

    for {
        select {
        case <- c.button_pressed:
        case <- c.floor_reached:
        case <- c.stop_button:
        case <- c.obstruction:
        case <- c.incoming:

        case <- time_to_ping.C:
            var p network.OutgoingPacket
            p.Data = []byte("Ping")
            c.outgoing_all <- p
        }
    }

}

func Client(c Channels, master network.ID) {
    log.Println("Starting client")
    // var local_queue Queue

    // for {
    //     select {
    //     case b := <- c.button_pressed:
    //     case f := <- c.floor_reached:
    //     case s := <- c.stop_button:
    //     case o := <- c.obstruction:
    //     case p, q := <- c.incoming:
    //         local_queue = i.Queue
    //     }
    // }
}

func TestNetwork(channels Channels) {
    ticker := time.NewTicker(5*time.Second)
    ticker_all := time.NewTicker(2*time.Second)
    for {
        select {
        case <- ticker.C:
            var p network.OutgoingPacket
            remote, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
            // remote, err := net.ResolveUDPAddr("udp", "78.91.16.212:20012")
            if err != nil {
                log.Fatal(err)
            }
            p.Destination = remote
            p.Data = []byte("Hello you!")
            channels.outgoing <- p

        case <- ticker_all.C:
            var p network.OutgoingPacket
            p.Data = []byte("Hello all!")
            channels.outgoing_all <- p

        case p := <- channels.incoming:
            log.Println("Got", len(p.Data), "bytes from", p.Sender)
        }
    }
}

func TestDriver(channels Channels) {
    for {
        channels.button_pressed <- driver.OrderButton{3, driver.ButtonDown}
        channels.button_pressed <- driver.OrderButton{4, driver.ButtonUp}
        channels.floor_reached <- 2
        time.Sleep(1 * time.Second)
    }
}

func main() {
    var listen_port int
    var start_as_master bool
    flag.IntVar(&listen_port, "port", 12345, "Preferred listen port")
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels Channels
    channels.button_pressed = make(chan driver.OrderButton)
    channels.floor_reached  = make(chan int)
    channels.stop_button    = make(chan bool)
    channels.obstruction    = make(chan bool)
    channels.incoming       = make(chan network.IncomingPacket)
    channels.outgoing       = make(chan network.OutgoingPacket)
    channels.outgoing_all   = make(chan network.OutgoingPacket)

    go driver.Init(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go network.Init(
        listen_port,
        channels.outgoing,
        channels.outgoing_all,
        channels.incoming)

    // go TestDriver(channels)

    if start_as_master {
        WaitForBackup(channels)
    } else {
        WaitForMaster(channels, nil)
    }

    // TestNetwork(channels)
}
