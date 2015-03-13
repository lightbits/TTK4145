package main

import (
    "./network"
    "./fakedriver"
)

type OrderType int32
const (
    OrderUp  = 0
    OrderDown = 1
    OrderOut  = 2
)

func WaitForBackup() {

}

func Master(backup network.ID) {

}

type Channels struct {
    button_pressed  chan driver.ButtonEvent
    floor_reached   chan driver.ReachedFloorEvent
    stop_button     chan driver.StopButtonEvent
    obstruction     chan driver.ObstructionEvent
    incoming        chan network.IncomingPacket
    outgoing        chan network.OutgoingPacket
    outgoing_all    chan network.OutgoingPacket
}

type Order struct {
    Floor    int
    Type     driver.ButtonType
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type Queue struct {
    Orders []Order
}

func (q Queue) RemoveOrdersAtFloor(f int) {
    for _, o := range(q.Orders) {
        if o.Floor == f {

        }
    }
}

func WaitForMaster(c Channels, q Queue) {

    // for {
    //     select {
    //     case b := <- c.button_pressed:
    //     case f := <- c.floor_reached:
    //     case s := <- c.stop_button:
    //     case o := <- c.obstruction:
    //     case i := <- c.incoming:
    //     }
    // }

}

func Client(c Channels) {
    // var local_queue Queue
}

func main() {
    var channels Channels
    channels.button_pressed = make(chan driver.ButtonEvent)
    channels.floor_reached  = make(chan driver.ReachedFloorEvent)
    channels.stop_button    = make(chan driver.StopButtonEvent)
    channels.obstruction    = make(chan driver.ObstructionEvent)
    channels.incoming       = make(chan network.IncomingPacket)
    channels.outgoing       = make(chan network.OutgoingPacket)
    channels.outgoing_all   = make(chan network.OutgoingPacket)

    go driver.Init(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go network.Init(
        12345,
        channels.outgoing,
        channels.outgoing_all,
        channels.incoming)
}
