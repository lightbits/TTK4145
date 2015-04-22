/*
TODO:
* Split up into modules: queue, client, master
* Add timer that fires when a client has not performed his order in a while
* Make client and master more clean
* Implement WaitForMaster
*/

package main

import (
    "flag"
    "./lift"
    "./network"
    "./driver"
    "./com"
    "./client"
    "./master"
)

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels com.Channels

    // Lift events
    channels.LastPassedFloorChanged = make(chan int)
    channels.TargetFloorChanged     = make(chan int)
    channels.CompletedFloor         = make(chan int)

    // Driver events
    channels.ButtonPressed  = make(chan driver.OrderButton)
    channels.FloorReached   = make(chan int)
    channels.StopButton     = make(chan bool)
    channels.Obstruction    = make(chan bool)

    // Network events
    channels.ToMaster       = make(chan network.Packet)
    channels.ToClients      = make(chan network.Packet)
    channels.FromMaster     = make(chan network.Packet)
    channels.FromClient     = make(chan network.Packet)

    driver.Init()

    go driver.Poll(
        channels.ButtonPressed,
        channels.FloorReached,
        channels.StopButton,
        channels.Obstruction)

    go lift.Init(
        channels.FloorReached,
        channels.LastPassedFloorChanged,
        channels.TargetFloorChanged,
        channels.CompletedFloor,
        channels.StopButton,
        channels.Obstruction)

    if start_as_master {
        go network.MasterWorker(channels.FromClient, channels.ToClients)
        go master.WaitForBackup(channels, nil, nilo)
    }

    go network.ClientWorker(channels.FromMaster, channels.ToMaster)
    client.WaitForMaster(channels, nil)
}
