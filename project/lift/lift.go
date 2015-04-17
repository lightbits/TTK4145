package lift

import (
    "time"
    "log"
    "../fakedriver"
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
        Startup
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
                    // TODO: if lpf == tf
                    fmt.Println("[LIFT]\tCompleted floor @ DoorOpen")
                    completed_floor <- target_floor
                    target_floor = -1
                    driver.CloseDoor()
                    state = Idle
                case Idle:    log.Fatal("Door closed @ Idle")
                case Moving:  log.Fatal("Door closed @ Moving")
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
                    }
                case Idle:     log.Fatal("Reached target @ Idle")
                case DoorOpen: log.Fatal("Reached target @ DoorOpen")
            }

        case <- stop_button: // ignore
        case <- obstruction: // ignore
        }
    }
}
