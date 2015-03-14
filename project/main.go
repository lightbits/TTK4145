package main

import (
    "time"
    "log"
    "fmt"
    "flag"
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
    to_master       chan network.OutgoingPacket
    to_all_clients  chan network.OutgoingPacket
    to_any_master   chan network.OutgoingPacket
    from_master     chan network.IncomingPacket
    from_client     chan network.IncomingPacket
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
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.from_client:
            // if packet.Sender != machine_id {
            Master(c, packet.Sender)
            return
        }
    }
}


func Master(c Channels, backup network.ID) {
    fmt.Println("[MASTER]\tStarting master with backup", backup)
    time_to_send := time.NewTicker(2*time.Second)
    for {
        select {
        case packet := <- c.from_client:
            fmt.Println("[MASTER]\tClient said", string(packet.Data))
        case <- time_to_send.C:
            c.to_all_clients <- network.OutgoingPacket {
                Data: []byte("This is an update from your master!"),
            }
        }
    }
}

func Backup(c Channels) {

}

func WaitForMaster(c Channels, remaining_orders []Order) {
    fmt.Println("[CLIENT]\tWaiting for master...")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case <- c.button_pressed:
        case <- c.floor_reached:
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tHeard a master!")
            Client(c, packet.Sender)
            return

        case <- time_to_ping.C:
            c.to_any_master <- network.OutgoingPacket {
                Data: []byte("Ping"),
            }

        case <- c.stop_button: // ignore
        case <- c.obstruction: // ignore
        }

    }

}

func Client(c Channels, master network.ID) {
    fmt.Println("[CLIENT]\tStarting client")
    // var local_queue Queue

    for {
        select {
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tMaster said", string(packet.Data))

        // case <- c.completed_floor:

        // case button := <- c.button_pressed:
        // case floor := <- c.floor_reached:
        //     if floor == target_floor {
        //         c.reached_target <- true
        //     }
        // case stopped := <- c.stop_button:
        // case obstructed := <- c.obstruction:

        }
    }
}

func TestNetwork(channels Channels) {
    t1 := time.NewTimer(1*time.Second)
    t2 := time.NewTimer(2*time.Second)
    t3 := time.NewTimer(3*time.Second)
    for {
        select {
        case <- t1.C:
            fmt.Println("Sending to all clients")
            channels.to_all_clients <- network.OutgoingPacket{
                Data: []byte("A")}

        case <- t2.C:
            fmt.Println("Sending to any master")
            channels.to_any_master <- network.OutgoingPacket{
                Data: []byte("BB")}

        case <- t3.C:
            fmt.Println("Sending to master")
            channels.to_master <- network.OutgoingPacket{
                Destination: "127.0.0.1:20012",
                Data: []byte("CCC")}

        case p := <- channels.from_client:
            log.Println("Client sent:", len(p.Data), "bytes from", p.Sender)

        case p := <- channels.from_master:
            log.Println("Master sent:", len(p.Data), "bytes from", p.Sender)
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
    channels.to_master       = make(chan network.OutgoingPacket)
    channels.to_all_clients  = make(chan network.OutgoingPacket)
    channels.to_any_master   = make(chan network.OutgoingPacket)
    channels.from_master     = make(chan network.IncomingPacket)
    channels.from_client     = make(chan network.IncomingPacket)

    go driver.Init(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go network.Init(
        listen_port,
        channels.to_master,
        channels.to_all_clients,
        channels.to_any_master,
        channels.from_master,
        channels.from_client)

    go lift.Init(
        channels.completed_floor,
        channels.reached_target)

    // TestNetwork(channels)
    if start_as_master {
        // The master also runs the client routine concurrently
        go WaitForMaster(channels, nil)
        WaitForBackup(channels)
    } else {

        // This is terrible. And it does not fix the problem that is
        // yet to come. Namely, when the backup should take over as master.
        // Then it needs to stop absorbing these.
        go func(from_client chan network.IncomingPacket) {
            for {
                <- from_client:
                fmt.Println("Heard an echo")
            }
        }(channels.from_client)
        WaitForMaster(channels, nil)
    }
}
