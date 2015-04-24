package main

import (
    "flag"
    "./lift"
    "./network"
    "./fakedriver"
    "./com"
    "./client"
    "./master"
    "os"
    "os/signal"
    "log"
)

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var lift_events com.LiftEvents
    lift_events.NewOrders      = make(chan []com.Order)
    lift_events.FloorReached   = make(chan int)
    lift_events.StopButton     = make(chan bool)
    lift_events.Obstruction    = make(chan bool)

    var client_events com.ClientEvents
    client_events.CompletedFloor = make(chan int)
    client_events.MissedDeadline = make(chan bool)
    client_events.ButtonPressed  = make(chan driver.OrderButton)
    client_events.ToMaster       = make(chan network.Packet)
    client_events.FromMaster     = make(chan network.Packet)

    var master_events com.MasterEvents
    master_events.ToClients  = make(chan network.Packet)
    master_events.FromClient = make(chan network.Packet)

    driver.Init()

    go driver.Poll(
        client_events.ButtonPressed,
        lift_events.FloorReached,
        lift_events.StopButton,
        lift_events.Obstruction)

    go lift.Init(
        client_events.CompletedFloor,
        client_events.MissedDeadline,
        lift_events.FloorReached,
        lift_events.NewOrders,
        lift_events.StopButton,
        lift_events.Obstruction)

    go network.ClientWorker(
        client_events.FromMaster,
        client_events.ToMaster)

    // Handle ctrl+c :)
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt)
    go func() {
        <- c
        driver.MotorStop()
        log.Fatal("[FATAL]\tUser terminated program")
    }()

    if start_as_master {
        go master.WaitForBackup(master_events, nil, nil)
    }

    client.WaitForMaster(client_events, master_events, lift_events, nil)
}
