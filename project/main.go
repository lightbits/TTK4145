package main

import (
    "fmt"
    "time"
)

// #cgo LDFLAGS: -L./driver -lelev -lcomedi -lm
// #include "driver/elev.h"
import "C"

func ElevInit() {
    C.elev_init()
}

func main() {
    C.elev_init()
    C.elev_set_motor_direction(-1)
    time.Sleep(1 * time.Second)
    C.elev_set_motor_direction(0)

    // for {
    //     var floor C.int = C.elev_get_floor_sensor_signal()
    //     if floor == 3 {
    //         C.elev_set_motor_direction(-1)
    //     } else if floor == 0 {
    //         C.elev_set_motor_direction(1)
    //     }
    // }
    fmt.Println("Hey!")
}