package driver

import (
    "log"
)

// TODO: Make an IO abstraction!!

// #include "io.h"
// #cgo LDFLAGS: -L. -lcomedi -lm
import "C"

const N_FLOORS = 4

type button_type int32
const (
    button_up   = 0
    button_down = 1
    button_out  = 2
)

func is_button_pressed(btn button_type, floor int) bool {
    if (floor < 0 || floor >= N_FLOORS) {
        return false
    }

    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},

    if (C.io_read_bit())
}

func DriverTest() {
    // Poll buttons and stuff




    // assert(floor >= 0);
    // assert(floor < N_FLOORS);
    // assert(!(button == BUTTON_CALL_UP && floor == N_FLOORS - 1));
    // assert(!(button == BUTTON_CALL_DOWN && floor == 0));
    // assert(button == BUTTON_CALL_UP || button == BUTTON_CALL_DOWN || button == BUTTON_COMMAND);

    // if (io_read_bit(button_channel_matrix[floor][button]))
    //     return 1;
    // else
    //     return 0;

    // // Poll floor sensors
    // if (io_read_bit(SENSOR_FLOOR1))
    //     return 0;
    // else if (io_read_bit(SENSOR_FLOOR2))
    //     return 1;
    // else if (io_read_bit(SENSOR_FLOOR3))
    //     return 2;
    // else if (io_read_bit(SENSOR_FLOOR4))
    //     return 3;
    // else
    //     return -1;

    // return io_read_bit(OBSTRUCTION);
    // return io_read_bit(STOP);
}

// type ButtonEvent struct {
//     Floor int
// }

// type SensorEvent struct {
//     Floor int
// }

// func Init(button_events chan ButtonEvent,
//           sensor_events chan SensorEvent) {

// }