package main

import (
    "time"
    "net"
    "log"
    "./network"
    "./fakedriver"
)

type OrderType int32
const (
    OrderUp  = 0
    OrderDown = 1
    OrderOut  = 2
)

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
    // TODO: Implement
    // for _, o := range(q.Orders) {
    //     if o.Floor == f {

    //     }
    // }
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

func WaitForBackup() {
}

func Master(backup network.ID) {

}

func Backup() {

}

func WaitForMaster(c Channels, q Queue) {

    // for {
    //     select {
    //     case b := <- c.button_pressed:
    //         // ignore
    //     case f := <- c.floor_reached:
    //     case s := <- c.stop_button:
    //     case o := <- c.obstruction:
    //     case i := <- c.incoming:
    //         if i.Protocol == MASTER_UPDATE:
    //             go Client(c)
    //     }
    // }

}

func Client(c Channels) {
    // var local_queue Queue

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
        55554,
        channels.outgoing,
        channels.outgoing_all,
        channels.incoming)

    ticker := time.NewTicker(1*time.Second)
    ticker_all := time.NewTicker(2*time.Second)
    for {
        select {
        case <- ticker.C:
            var p network.OutgoingPacket
            remote, err := net.ResolveUDPAddr("udp", "78.91.19.229:20012")
            // remote, err := net.ResolveUDPAddr("udp", "78.91.16.212:12345")
            if err != nil {
                log.Fatal(err)
            }
            p.Destination = remote
            p.Data = []byte("Hello you!")
            channels.outgoing <- p

        case <- ticker_all.C:
            var p network.OutgoingPacket
            p.Data = []byte("Hello all!")
            channels.outgoing_all <- p

        case p := <- channels.incoming:
            log.Println("Got", len(p.Data), "bytes from", p.Sender)
        }
    }
}
