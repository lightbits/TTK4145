package main

import (
    "log"
    "./driver"
)

type OrderType int32
const (
    OrderUp  = 0
    OrderDown = 1
    OrderOut  = 2
)

func Master() {
    // type network.LiftID struct {
    //     IPAddress uint32
    //     Port      uint16
    // }

    // type network.ClientOrder struct {
    //     Floor int
    //     Type  OrderType
    //     Done  bool
    // }

    // type network.ClientUpdate struct {
    //     ID              network.LiftID
    //     PendingUpdates  []network.ClientOrder
    //     LastPassedFloor int
    // }

    // type network.MasterOrder struct {
    //     Floor      int
    //     Type       OrderType
    //     AssignedTo network.LiftID
    // }

    // type network.MasterUpdate struct {
    //     OrderList []network.MasterOrder
    // }

    // /*
    // Wait a sec.
    // What if we make these interfaces?
    // And they must support a .ConvertToByteBlob() function?
    // That might be cleaner...
    // */

    // incoming_client_update := make(chan network.ClientUpdate)
    // outgoing_master_update := make(chan network.MasterUpdate)
    // go network.Init(incoming_client_update, nil, outgoing_master_update, nil)

    // type order struct {
    //     Floor      int
    //     Type       order_type
    //     AssignedTo lift_id
    // }

    // type client struct {
    //     ID           lift_id
    //     TimeoutTimer time.Timer
    //     TimedOut     bool
    // }

    // type client_update struct {
    //     ID lift_id

    // }

    // order_list     []order
    // lift_positions map[lift_id]int
    // active_clients []client

    // for {
    //     select {
    //     case update := <- incoming_client_update:
    //         if
    //         for _, request := update.Requests: {
    //         }
    //     }
    //     case <- master_update_timer:
    //         outgoing_master_update <- network.MasterUpdate{order_list}

    //     for c := range active_clients {
    //         case <- c.TimeoutTimer:
    //             c.TimedOut = true
    //     }
    // }
}

func Client() {
    // type order struct {
    //     Floor     int
    //     Type      order_type
    //     TakenBy   lift_id
    //     Finished  bool
    // }

    // go driver.Init(button_pressed, floor_reached, stop_pressed, obstruction)
    // go lift.Init(finished_order)
    // for {
    //     select {
    //     case button := <- button_pressed:
    //         log.Println("Button ", button.Type, button.Floor)

    //     case floor := <- floor_reached:
    //         client_status.last_passed_floor = floor
    //         log.Println("Floor ", floor.FloorIndex)

    //     case update := <- master_update:
    //         pending_orders = update.pending_orders

    //     case floor := <- lift_finished_order:
    //         // After the lift closes the doors after three seconds
    //         // after reaching a floor

    //     case <- client_update_timer:
    //         client_update <- client_status

    //     case <- master_timeout_timer:
    //         /*
    //         if we_are_backup {
    //             BecomePrimary()
    //         } else {
    //             Wait until we hear from a master. Do not accept any input.
    //             WaitForMaster()
    //         }
    //         */

    //     // Don't care
    //     case <- obstruction:
    //         log.Println("Obstruction was toggled")

    //     // Don't care
    //     case <- stop_pressed:
    //         log.Println("Stop was pressed")
    //     }
    // }
}

func main() {
    button_pressed := make(chan driver.ButtonEvent)
    floor_reached  := make(chan driver.ReachedFloorEvent)
    stop_pressed   := make(chan driver.StopButtonEvent)
    obstruction    := make(chan driver.ObstructionEvent)

    go driver.Init(button_pressed, floor_reached, stop_pressed, obstruction)

    for {
        select {
        case btn := <- button_pressed:
            if btn.Floor == 0 {
                driver.MotorUp()
                driver.SetButtonLamp(driver.ButtonOut, 0, true)
                driver.SetButtonLamp(driver.ButtonOut, 1, true)
                driver.SetButtonLamp(driver.ButtonOut, 2, true)
                driver.SetButtonLamp(driver.ButtonOut, 3, true)
                driver.SetButtonLamp(driver.ButtonUp, 0, true)
                driver.SetButtonLamp(driver.ButtonUp, 1, true)
                driver.SetButtonLamp(driver.ButtonUp, 2, true)
                driver.SetButtonLamp(driver.ButtonDown, 1, true)
                driver.SetButtonLamp(driver.ButtonDown, 2, true)
                driver.SetButtonLamp(driver.ButtonDown, 3, true)
                driver.SetDoorOpenLamp(true)
                driver.SetStopLamp(true)
            } else if btn.Floor == 1 {
                driver.MotorDown()
            } else if btn.Floor == 2 {
                driver.MotorStop()
            }
        case floor := <- floor_reached:
            driver.SetFloorIndicator(floor.FloorIndex)
            log.Println(floor.FloorIndex)
        case <- stop_pressed:
            log.Println("Stop")
        case <- obstruction:
            log.Println("Obstruction")
        }
    }

    // finished_order := make(chan lift.FinishedOrderEvent)
    // client_update  := make(chan network.ClientUpdate)
    // master_update  := make(chan network.MasterUpdate)
    // master_timeout_timer := time.NewTimer()
    // client_update_timer := time.NewTimer()

    log.Println("Hello!")
}
