package lift

import (
    "time"
    "../driver"
)

func Init(
    floor_reached             chan int,
    last_passed_floor_changed chan int,
    new_floor_order           chan int,
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
                    completed_floor <- target_floor
                    driver.CloseDoor()
                    state = Idle
                case Idle:    // Ignoring
                case Moving:  // Ignoring
            }

        case floor := <- new_floor_order:
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
                case Idle:     // Ignoring
                case DoorOpen: // Ignoring
            }

        case <- stop_button: // Ignoring
        case <- obstruction: // Ignoring
        }
    }
}
