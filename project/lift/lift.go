package lift

import (
    "time"
    "../driver"
    "../logger"
)

var last_passed_floor int

func GetLastPassedFloor() int {
    return last_passed_floor
}

func StatemachineLoop(
    completed_floor  chan int,
    missed_deadline  chan bool,
    floor_reached    chan int,
    new_target_floor chan int,
    stop_button      chan bool,
    obstruction      chan bool) {

    ORDER_DEADLINE_INTERVAL := 5 * driver.N_FLOORS * time.Second
    deadline_timer := time.NewTimer(ORDER_DEADLINE_INTERVAL)
    deadline_timer.Stop()

    door_timer := time.NewTimer(3 * time.Second)
    door_timer.Stop()

    type State int
    const (
        Idle State = iota
        DoorOpen
        Moving
    )
    state := Idle

    last_passed_floor = 0
    target_floor := driver.INVALID_FLOOR

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case DoorOpen:
                    println(logger.Info, "Door timer @ DoorOpen")
                    driver.CloseDoor()
                    state = Idle
                    completed_floor <- target_floor
                    target_floor = driver.INVALID_FLOOR
                    deadline_timer.Stop()
                case Idle:    println(logger.Debug, "Door timer @ Idle")
                case Moving:  println(logger.Debug, "Door timer @ Moving")
            }

        case <- deadline_timer.C:
            missed_deadline <- true

        case floor := <- new_target_floor:
            if target_floor != floor {
                deadline_timer.Reset(ORDER_DEADLINE_INTERVAL)
            }
            target_floor = floor
            switch (state) {
                case Idle:
                    println(logger.Info, "New order @ Idle")
                    if target_floor == driver.INVALID_FLOOR {
                        break
                    } else if target_floor > last_passed_floor {
                        state = Moving
                        driver.MotorUp()
                    } else if target_floor < last_passed_floor {
                        state = Moving
                        driver.MotorDown()
                    } else {
                        door_timer.Reset(3 * time.Second)
                        driver.OpenDoor()
                        driver.MotorStop()
                        state = DoorOpen
                    }
                case Moving:   println(logger.Debug, "New order @ Moving")
                case DoorOpen: println(logger.Debug, "New order @ DoorOpen")
            }

        case floor := <- floor_reached:
            last_passed_floor = floor
            switch (state) {
                case Moving:
                    println(logger.Info, "Reached floor", floor, "@ Moving")
                    driver.SetFloorIndicator(floor)
                    if target_floor == driver.INVALID_FLOOR {
                        break
                    } else if target_floor > last_passed_floor {
                        state = Moving
                        driver.MotorUp()
                    } else if target_floor < last_passed_floor {
                        state = Moving
                        driver.MotorDown()
                    } else {
                        door_timer.Reset(3 * time.Second)
                        driver.OpenDoor()
                        driver.MotorStop()
                        state = DoorOpen
                    }
                case Idle:     println(logger.Info, "Reached floor", floor, "@ Idle")
                case DoorOpen: println(logger.Info, "Reached floor", floor, "@ DoorOpen")
            }

        case <- stop_button: // Ignoring
        case <- obstruction: // Ignoring
        }
    }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "LIFT", args...)
}
