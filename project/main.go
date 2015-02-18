package main

import (
    "log"
    "encoding/binary"
    "bytes"
    // "./driver"
)

type order_type int32
const (
    order_up   = 1
    order_out  = 0
    order_down = -1
)

type lift_id struct {
    IPAddress uint32
    Port      uint16
}

type order struct {
    Floor     int
    Type      order_type
    TakenBy   lift_id
    Finished  bool
}

type client_update struct {
    Requests   []order
    Requesting int
}

type master_update struct {
    PendingOrders []order
}

func main() {
    // button_pressed  := make(chan driver.ButtonEvent)
    // floor_reached   := make(chan driver.FloorEvent)
    // stop_pressed    := make(chan driver.StopEvent)
    // obstruction     := make(chan driver.ObstructionEvent)

    // go driver.Init(button_pressed, floor_reached, stop_pressed, obstruction)
    // for {
    //     select {
    //     case button := <- button_pressed:
    //         log.Println("Button ", button.Type, button.Floor)
    //     case obstructed := <- obstruction:
    //         log.Println("Obstructed ", obstructed)
    //     case <- stop_pressed:
    //         log.Println("Stop")
    //     case floor := <- floor_reached:
    //         log.Println("Floor ", floor)
    //     }
    // }

    OrderA := order{
        Floor: 0,
        Type: order_up,
        TakenBy: lift_id{0xabad1dea, 0xbeef},
    }

    OrderB := order{
        Floor: 1,
        Type: order_down,
        TakenBy: lift_id{0xaabababa, 0xbeef},
    }

    PendingOrders := []order{OrderA, OrderB}
    Update := master_update{PendingOrders}
    for _, Order := range(Update.PendingOrders) {
        PrintOrder(Order)
    }

    log.Println("Hello!")
}
