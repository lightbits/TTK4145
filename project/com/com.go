package com

import (
    "../driversim"
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

type LiftEvents struct {
    FloorReached    chan int
    NewTargetFloor  chan int
    StopButton      chan bool
    Obstruction     chan bool
}

type ClientEvents struct {
    CompletedFloor chan int
    MissedDeadline chan bool
    ButtonPressed  chan driver.OrderButton
    FromMaster     chan network.Packet
    ToMaster       chan network.Packet
}

type MasterEvents struct {
    ToClients  chan network.Packet
    FromClient chan network.Packet
}
