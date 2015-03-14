package lift

import (
    "time"
    "log"
)

// TODO: Stop button event, obstruction event. Belong here
// or in event manager?
func Init(
    completed_floor chan bool,
    reached_target  chan bool) {

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

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case DoorOpen:
                    completed_floor <- true
                    state = Idle
                case Startup: log.Fatal("Door closed @ Startup")
                case Idle:    log.Fatal("Door closed @ Idle")
                case Moving:  log.Fatal("Door closed @ Moving")
            }

        case <- reached_target:
            switch (state) {
                case Moving:
                    door_timer.Reset(3 * time.Second)
                    state = DoorOpen
                case Startup:
                    state = Idle
                case Idle:     log.Fatal("Reached target @ Idle")
                case DoorOpen: log.Fatal("Reached target @ DoorOpen")
            }

        // These can not be handled both in main and in here...
        // case <- stop_button: // ignore
        // case <- obstruction: // ignore
        }
    }
}
