package lift

import (
    "time"
    "log"
    "../driver"
    "fmt"
)

func Init(
    floor_reached             chan int,
    last_passed_floor_changed chan int,
    target_floor_changed      chan int,
    completed_floor           chan int,
    stop_button               chan bool,
    obstruction               chan bool) {

    door_timer := time.NewTimer(3 * time.Second)
    door_timer.Stop()
    type State int
    const (
        Idle State = iota
        DoorOpen
        Moving
    )
    state := Idle

    last_passed_floor := 0
    target_floor := 0

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case DoorOpen:
                    // TODO: if lpf == tf, incase lift was dragged
                    fmt.Println("[LIFT]\tCompleted floor @ DoorOpen")
                    completed_floor <- target_floor
                    // target_floor = -1 // TODO: Remove this?
                    driver.CloseDoor()
                    state = Idle
                case Idle:    log.Println("Door closed @ Idle")
                case Moving:  log.Println("Door closed @ Moving")
            }

        case floor := <- target_floor_changed:
            target_floor = floor
            switch (state) {
                case Idle:
                    if target_floor > last_passed_floor {
                        state = Moving
                        driver.MotorUp()
                    } else if target_floor < last_passed_floor {
                        state = Moving
                        driver.MotorDown()
                    } else {
                        door_timer.Reset(3 * time.Second)
                        driver.OpenDoor()
                        state = DoorOpen
                    }
                case Moving:   // Ignoring
                case DoorOpen: // Ignoring
            }

        case floor := <- floor_reached:
            last_passed_floor = floor
            last_passed_floor_changed <- floor
            switch (state) {
                case Moving:
                    driver.SetFloorIndicator(floor)
                    if floor == target_floor {
                        door_timer.Reset(3 * time.Second)
                        driver.MotorStop()
                        driver.OpenDoor()
                        state = DoorOpen
                    } else if target_floor > floor {
                        driver.MotorUp()
                    } else if target_floor < floor {
                        driver.MotorDown()
                    }
                case Idle:     log.Println("Reached floor @ Idle")
                case DoorOpen: log.Println("Reached floor @ DoorOpen")
            }

        case <- stop_button: // ignore
        case <- obstruction: // ignore
        }
    }
}
