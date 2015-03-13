package lift

import (
    "time"
    "log"
)

func Init(
    completed_floor chan bool,
    reached_target  chan bool,
    stop_button     chan bool,
    obstruction     chan bool) {

    door_timer := time.NewTimer(3 * time.Second)
    door_timer.Stop()
    type State int
    const (
        Idle State = iota
        Startup
        DoorOpen
        Moving
    )
    state := Startup

    select {
    case <- reached_target:
        state = Idle
    }

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case DoorOpen:
                    completed_floor <- true
                case Idle:   log.Fatal("Door closed @ Idle")
                case Moving: log.Fatal("Door closed @ Moving")
            }

        case <- reached_target:
            switch (state) {
                case Moving:
                    state = DoorOpen
                    door_timer.Reset(3 * time.Second)
                case Idle:     log.Fatal("Reached target @ Idle")
                case DoorOpen: log.Fatal("Reached target @ DoorOpen")
            }

        case <- stop_button: // ignore
        case <- obstruction: // ignore
        }
    }
}
