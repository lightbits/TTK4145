package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "./lift"
    "./network"
    "./client"
    "./master"
    "./queue"
    "./fakedriver"
)

func TestNetwork(
    c network.ClientEvents,
    m network.MasterEvents) {

    go network.ClientWorker(c)
    go network.MasterWorker(m)
    t1 := time.NewTimer(1*time.Second)
    t2 := time.NewTimer(2*time.Second)

    for {
        select {
        case <- t1.C:
            fmt.Println("[NETTEST]\tSending to all clients")
            m.to_clients <- network.Packet{
                Data: []byte("A")}

        case <- t2.C:
            fmt.Println("[NETTEST]\tSending to any master")
            c.to_master <- network.Packet{
                Data: []byte("BB")}

        case p := <- m.from_client:
            fmt.Println("[NETTEST]\tClient sent:", len(p.Data), "bytes from", p.Address)

        case p := <- c.from_master:
            fmt.Println("[NETTEST]\tMaster sent:", len(p.Data), "bytes from", p.Address)
        }
    }
}

func TestDriver(e driver.Events) {
    for {
        select {
        case btn := <- e.button_pressed:
            fmt.Println("[TEST]\tButton pressed")
            driver.SetButtonLamp(btn, true)
        case <- e.floor_reached:
            fmt.Println("[TEST]\tFloor reached")
        case <- e.stop_button:
            driver.ClearAllButtonLamps()
            fmt.Println("[TEST]\tStop button pressed")
        case <- e.obstruction:
            fmt.Println("[TEST]\tObstruction changed")
        }
    }
}

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var lift_events                       lift.Events
    lift_events.last_passed_floor_changed := make(chan int)
    lift_events.target_floor_changed      := make(chan int)
    lift_events.completed_floor           := make(chan int)

    var io_events                         driver.Events
    io_events.button_pressed              := make(chan driver.OrderButton)
    io_events.floor_reached               := make(chan int)
    io_events.stop_button                 := make(chan bool)
    io_events.obstruction                 := make(chan bool)

    var net_client_events                 network.ClientEvents
    net_client_events.to_master           := make(chan network.ClientData)
    net_client_events.from_master         := make(chan network.MasterData)

    var net_master_events                 network.MasterEvents
    net_master_events.from_client         := make(chan network.ClientData)
    net_master_events.to_clients          := make(chan network.MasterData)

    driver.Init()
    // OBS this passes in button pressed, which is not used in lift...
    go lift.Init(lift_events, io_events)
    go driver.Poll(io_events)
    go network.ClientWorker(net_client_events)

    if start_as_master {
        initial_queue := make([]Order, 0)
        go WaitForBackup(net_master_events, initial_queue)
    }

    WaitForMaster(lift_events, io_events, net_client_events, nil)
}
