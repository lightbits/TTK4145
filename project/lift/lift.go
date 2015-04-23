package lift

import (
    "time"
    "../fakedriver"
    "../com"
    "../queue"
    "../network"
    "fmt"
)

var last_passed_floor int

func GetLastPassedFloor() int {
    return last_passed_floor
}

func Init(
    floor_reached   chan int,
    completed_floor chan int,
    new_orders      chan []com.Order,
    missed_deadline chan bool,
    stop_button     chan bool,
    obstruction     chan bool) {

    last_passed_floor = 0

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

    last_passed_floor := 0
    target_floor := 0

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case DoorOpen:
                    completed_floor <- target_floor
                    deadline_timer.Stop()
                    driver.CloseDoor()
                    state = Idle
                case Idle:    // Ignoring
                case Moving:  // Ignoring
            }

        case <- deadline_timer.C:
            missed_deadline <- true

        case orders := <- new_orders:
            target_floor = queue.GetPriority(orders, network.GetMachineID())
            if target_floor != driver.INVALID_FLOOR {
                deadline_timer.Reset(ORDER_DEADLINE_INTERVAL)
            }
            switch (state) {
                case Idle:
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
                        state = DoorOpen
                    }
                case Moving:   // Ignoring
                case DoorOpen: // Ignoring
            }

        case floor := <- floor_reached:
            last_passed_floor = floor
            fmt.Println(last_passed_floor)
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
