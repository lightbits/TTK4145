package com

import (
    "../fakedriver"
    "../network"
    "time"
    "encoding/json"
    "log"
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
    Clients        map[network.ID]Client
}

type Client struct {
    ID              network.ID
    LastPassedFloor int
    HasTimedOut     bool
    AliveTimer      *time.Timer `json:"-"`
}

func DecodeClientPacket(b []byte) (ClientData, error) {
    var result ClientData
    err := json.Unmarshal(b, &result)
    return result, err
}

func EncodeMasterData(m MasterData) []byte {
    result, err := json.Marshal(m)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func DecodeMasterPacket(b []byte) (MasterData, error) {
    var result MasterData
    err := json.Unmarshal(b, &result)
    return result, err
}

func EncodeClientData(c ClientData) []byte {
    result, err := json.Marshal(c)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

/*
TODO:
type LiftEvents struct {
    LastPassedFloorChanged chan int
    NewFloorOrder          chan int
    CompletedFloor         chan int
    MissedDeadline         chan bool
}

type DriverEvents struct {
    ButtonPressed  chan driver.OrderButton
    FloorReached   chan int
    StopButton     chan bool
    Obstruction    chan bool
}

// Only one, since the client needs both types!
type NetworkEvents {
    ToMaster   chan network.Packet
    FromMaster chan network.Packet
    ToClients  chan network.Packet
    FromClient chan network.Packet
}

*/

type Channels struct {
    // Lift events
    NewFloorOrder  chan int
    CompletedFloor chan int
    MissedDeadline chan bool
    NewOrders      chan []Order

    // Driver events
    ButtonPressed  chan driver.OrderButton
    FloorReached   chan int
    StopButton     chan bool
    Obstruction    chan bool

    // Network events (client-side)
    ToMaster       chan network.Packet
    FromMaster     chan network.Packet

    // Network events (master-side)
    ToClients      chan network.Packet
    FromClient     chan network.Packet
}
