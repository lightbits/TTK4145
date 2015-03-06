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
