package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "./lift"
    "./network"
    "./driver"
    // "./fakedriver"
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
    AssignedBackup network.ID
    Orders         []Order
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
    to_master       chan network.Packet
    to_clients      chan network.Packet
    from_master     chan network.Packet
    from_client     chan network.Packet
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

func WaitForBackup(c Channels, initial_queue []Order) {
    go network.MasterWorker(c.from_client, c.to_clients)
    machine_id := network.GetMachineID()
    fmt.Println("[MASTER]\tRunning on machine", machine_id)
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.from_client:
            if packet.Address != machine_id {
                Master(c, packet.Address)
                return
            } else {
                fmt.Println("[MASTER]\tCannot use own machine as backup client")
            }
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
            c.to_clients <- network.Packet {
                Data: []byte("This is an update from your master!"),
            }
        }
    }
}

func WaitForMaster(c Channels, remaining_orders []Order) {
    fmt.Println("[CLIENT]\tWaiting for master...")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tHeard a master!")
            Client(c, packet.Address)
            return

        case <- time_to_ping.C:
            c.to_master <- network.Packet {
                Data: []byte("Ping"),
            }
        case <- c.button_pressed:
        case <- c.floor_reached:
        case <- c.stop_button: // ignore
        case <- c.obstruction: // ignore
        }

    }

}

func Client(c Channels, master network.ID) {
    fmt.Println("[CLIENT]\tStarting client")
    is_backup := false
    master_timeout := time.NewTimer(5*time.Second)
    local_queue := make([]Order, 0)
    for {
        select {
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tMaster said", string(packet.Data))
        case <- master_timeout.C:
            if is_backup {
                WaitForBackup(c, local_queue)
            }
        // case <- c.completed_floor:
        // case button := <- c.button_pressed:
        // case floor := <- c.floor_reached:
        // case stopped := <- c.stop_button:
        // case obstructed := <- c.obstruction:
        }
    }
}

func TestNetwork(channels Channels) {
    go network.ClientWorker(channels.from_master, channels.to_master)
    go network.MasterWorker(channels.from_client, channels.to_clients)
    t1 := time.NewTimer(1*time.Second)
    t2 := time.NewTimer(2*time.Second)
    for {
        select {
        case <- t1.C:
            fmt.Println("[NETTEST]\tSending to all clients")
            channels.to_clients <- network.Packet{
                Data: []byte("A")}

        case <- t2.C:
            fmt.Println("[NETTEST]\tSending to any master")
            channels.to_master <- network.Packet{
                Data: []byte("BB")}

        case p := <- channels.from_client:
            fmt.Println("[NETTEST]\tClient sent:", len(p.Data), "bytes from", p.Address)

        case p := <- channels.from_master:
            fmt.Println("[NETTEST]\tMaster sent:", len(p.Data), "bytes from", p.Address)
        }
    }
}

func TestDriver(channels Channels) {
    for {
        select {
        case btn := <- channels.button_pressed:
            fmt.Println("[TEST]\tButton pressed")
            driver.SetButtonLamp(btn, true)
        case <- channels.floor_reached:
            fmt.Println("[TEST]\tFloor reached")
        case <- channels.stop_button:
            fmt.Println("[TEST]\tStop button pressed")
        case <- channels.obstruction:
            fmt.Println("[TEST]\tObstruction changed")
        }
    }
}

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels Channels
    channels.completed_floor = make(chan bool)
    channels.reached_target  = make(chan bool)
    channels.button_pressed  = make(chan driver.OrderButton)
    channels.floor_reached   = make(chan int)
    channels.stop_button     = make(chan bool)
    channels.obstruction     = make(chan bool)
    channels.to_master       = make(chan network.Packet)
    channels.to_clients      = make(chan network.Packet)
    channels.from_master     = make(chan network.Packet)
    channels.from_client     = make(chan network.Packet)

    driver.Init()

    go driver.Poll(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go lift.Init(
        channels.completed_floor,
        channels.reached_target)

    TestDriver(channels)
    // go network.ClientWorker(channels.from_master, channels.to_master)

    // if start_as_master {
    //     go WaitForMaster(channels, nil)
    //     WaitForBackup(channels, nil)
    // } else {
    //     WaitForMaster(channels, nil)
    // }
}
