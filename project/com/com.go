package com

import (
    "../driver"
    "../network"
    "../queue"
    "encoding/json"
    "log"
)

type ClientData struct {
    LastPassedFloor int
    Requests        []queue.Order
}

type MasterData struct {
    AssignedBackup network.ID
    Orders         []queue.Order
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

type Channels struct {
    // Lift events
    LastPassedFloorChanged chan int
    TargetFloorChanged     chan int
    CompletedFloor         chan int

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