package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "./network"
    "./fakedriver"
    "./lift"
)

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type Queue []Order

type ClientData struct {
    // Protocol int
    LastPassedFloor int
    Requests Queue
}

type MasterData struct {
    Orders Queue
}

type ClientType struct {
    ID              network.ID
    LastPassedFloor int
    TargetFloor     int
}

type Channels struct {
    completed_floor chan bool
    reached_target  chan bool
    button_pressed  chan driver.OrderButton
    floor_reached   chan int
    stop_button     chan bool
    obstruction     chan bool
    incoming        chan network.IncomingPacket
    outgoing        chan network.OutgoingPacket
    outgoing_all    chan network.OutgoingPacket
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
    fmt.Println("Waiting for backup...")
    for {
        select {
        case packet := <- c.incoming:
            Master(c, packet.Sender)
            return
        }
    }
}

func Master(c Channels, backup network.ID) {
    fmt.Println("Starting master with backup", backup)
    time_to_send := time.NewTicker(1*time.Second)
    for {
        select {
        case <- time_to_send.C:
            c.outgoing_all <- network.OutgoingPacket {
                Data: []byte("This is an update from your master!"),
            }
        }
    }
}

func Backup(c Channels) {

}

func SendPing(c chan network.OutgoingPacket) {
    c <- network.OutgoingPacket {
        Data: []byte("Ping"),
    }
}

func WaitForMaster(c Channels, q Queue) {
    log.Println("Waiting for master...")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case <- c.button_pressed:
        case <- c.floor_reached:
        case <- c.stop_button:
        case <- c.obstruction:

        // TODO: Should we have some protocol stuff?
        // Verify that we did infact get a packet from the master?
        // Might just want to include in ClientData and MasterData
        case packet := <- c.incoming:
            Client(c, packet.Sender)
            return

        case <- time_to_ping.C:
            SendPing(c.outgoing_all)
        }
    }

}

func Client(c Channels, master network.ID) {
    log.Println("Starting client")
    // var local_queue Queue

    // for {
    //     select {
    //     case <- c.completed_floor:

    //     // case button := <- c.button_pressed:
    //     case floor := <- c.floor_reached:
    //         if floor == target_floor {
    //             c.reached_target <- true
    //         }
    //     // case stopped := <- c.stop_button:
    //     // case obstructed := <- c.obstruction:
    //     // case packet := <- c.incoming:

    //     }
    // }
}

func TestNetwork(channels Channels) {
    ticker := time.NewTicker(5*time.Second)
    ticker_all := time.NewTicker(2*time.Second)
    for {
        select {
        case <- ticker.C:
            channels.outgoing <- network.OutgoingPacket{
                Destination: "127.0.0.1:20012",
                Data: []byte("Hi!")}

        case <- ticker_all.C:
            channels.outgoing_all <- network.OutgoingPacket{
                Data: []byte("hello!")}

        case p := <- channels.incoming:
            log.Println("Got", len(p.Data), "bytes from", p.Sender)
        }
    }
}

func main() {
    var listen_port int
    var bcast_port int
    var start_as_master bool
    flag.IntVar(&listen_port, "port", 12345, "Preferred listen port")
    flag.IntVar(&bcast_port, "bport", 20012, "Broadcast port")
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels Channels
    channels.completed_floor = make(chan bool)
    channels.reached_target  = make(chan bool)
    channels.button_pressed  = make(chan driver.OrderButton)
    channels.floor_reached   = make(chan int)
    channels.stop_button     = make(chan bool)
    channels.obstruction     = make(chan bool)
    channels.incoming        = make(chan network.IncomingPacket)
    channels.outgoing        = make(chan network.OutgoingPacket)
    channels.outgoing_all    = make(chan network.OutgoingPacket)

    go driver.Init(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go network.Init(
        listen_port,
        bcast_port,
        channels.outgoing,
        channels.outgoing_all,
        channels.incoming)

    go lift.Init(
        channels.completed_floor,
        channels.reached_target,
        channels.stop_button,
        channels.obstruction)

    if start_as_master {
        WaitForBackup(channels)
    } else {
        WaitForMaster(channels, nil)
    }
}
