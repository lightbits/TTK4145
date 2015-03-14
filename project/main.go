package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "encoding/json"
    "./lift"
    "./network"
    "./fakedriver"
)

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type ClientData struct {
    LastPassedFloor int
    Requests        []Order
}

type MasterData struct {
    Orders   []Order
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

        /*
        case packet := <- c.message_from_client:
            if packet.Sender != machine_id {
                Master(c, packet.Sender)
            }
        */
        }
    }
}


func Master(c Channels, backup network.ID) {
    fmt.Println("Starting master with backup", backup)
    time_to_send := time.NewTicker(1*time.Second)
    // orders := make([]Order, 0)
    for {
        select {
        case <- time_to_send.C:
            // SendOrdersToClients(c.outgoing_all, orders)
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

func WaitForMaster(c Channels, remaining_orders []Order) {
    fmt.Println("Waiting for master...")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case <- c.button_pressed:
        case <- c.floor_reached:

        /* TODO: Message protocols?
        To verify that the incoming packet is infact from the master
        */
        // TODO: Should we have some protocol stuff?
        // Verify that we did infact get a packet from the master?
        // Might just want to include in ClientData and MasterData
        // Otherwise, can we guranatee
        case packet := <- c.incoming:
            Client(c, packet.Sender)
            return

        case <- time_to_ping.C:
            SendPing(c.outgoing_all)

        case <- c.stop_button: // ignore
        case <- c.obstruction: // ignore
        }

    }

}

func Client(c Channels, master network.ID) {
    fmt.Println("Starting client")
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
    var start_as_master bool
    flag.IntVar(&listen_port, "port", 20012, "Port that all clients send and listen to")
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

    b := []byte{0x00, 0x00, 0x00, 0x01}
    i := int(b)
    fmt.Println(i)

    b = bytes.NewBuffer(packet.UserData[:packet.Length])
    r = request{}
    binary.Read(b, binary.BigEndian, &r)

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

    go lift.Init(
        channels.completed_floor,
        channels.reached_target)

    if start_as_master {
        go WaitForMaster(channels, nil) // Also launch client instance
        WaitForBackup(channels)
    } else {
        WaitForMaster(channels, nil)
    }
}
