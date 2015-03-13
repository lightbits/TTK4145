package lift

import (
    "time"
)

func StateMachine(reached_floor chan ReachedFloorEvent,
                  obstruction chan ObstructionEvent,
                  stop_button chan StopButtonEvent,
                  new_target_floor chan int,
                  cleared_floor chan ClearedFloorEvent) {

    door_timer := time.NewTimer(3 * timeSecond)

    for {
        select {
        case f := <- new_target_floor:

        case f := <- reached_floor:
        case o := <- obstruction:
            // ignore
        case s := <- stop_button:
            // ignore
        }
    }
}
