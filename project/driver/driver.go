package driver

import (
    "log"
    "time"
)

type ButtonType int
const (
    ButtonUp ButtonType = iota
    ButtonDown
    ButtonOut
)
type ButtonEvent struct {
    Floor int
    Type  ButtonType
}

type FloorEvent       int
type StopEvent        bool
type ObstructionEvent bool

type io_event struct {
    bit    int
    is_set bool
}

func poll(bits [N_FLOORS]bit_state, event chan io_event) {
    for {
        for i := 0; i < len(bits); i++ {
            is_set := io_read_bit(bits[i].channel) == 1

            if (!bits[i].was_set && is_set) {
                event <- io_event{i, true}
            } else if (bits[i].was_set && !is_set) {
                event <- io_event{i, false}
            }
            bits[i].was_set = is_set
        }

        // If we hammer the io module too much it actually fails to read.
        // So we sleep a little bit.
        time.Sleep(1 * time.Millisecond)
    }
}

type MotorDirection int
const (
    MOTOR_DIR_UP   = 1
    MOTOR_DIR_STOP = 0
    MOTOR_DIR_DOWN = -1
)

func SetMotorDirection(dir MotorDirection) {
    if (dir == 0){
        io_write_analog(MOTOR, 0);
    } else if (dir > 0) {
        io_clear_bit(MOTORDIR);
        io_write_analog(MOTOR, 2800);
    } else if (dir < 0) {
        io_set_bit(MOTORDIR);
        io_write_analog(MOTOR, 2800);
    }
}

func Init(button_pressed chan ButtonEvent,
          floor_reached  chan FloorEvent,
          stop_pressed   chan StopEvent,
          obstruction    chan ObstructionEvent) {

    if (!io_init()) {
        log.Fatal("Failed to initialize driver")
    }

    up_ch   := make(chan io_event)
    down_ch := make(chan io_event)
    out_ch  := make(chan io_event)
    flr_ch  := make(chan io_event)
    stp_ch  := make(chan io_event)
    obs_ch  := make(chan io_event)

    if (button_pressed != nil) {
        go poll(UP_BUTTONS, up_ch)
        go poll(DOWN_BUTTONS, down_ch)
        go poll(OUT_BUTTONS, out_ch)
    }

    if (floor_reached != nil) {
        go poll(FLOOR_SENSORS, flr_ch)
    }

    if (stop_pressed != nil) {
        go poll(STOP_BUTTONS, stp_ch)
    }

    if (obstruction != nil) {
        go poll(OBSTRUCTIONS, obs_ch)
    }

    for {
        select {
        case e := <-up_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, ButtonUp}
            }
        case e := <-down_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, ButtonDown}
            }
        case e := <-out_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, ButtonOut}
            }
        case e := <-flr_ch:
            if (e.is_set) {
                floor_reached <- FloorEvent(e.bit)
            }
        case e := <- stp_ch:
            if (e.is_set) {
                stop_pressed <- StopEvent(true)
            }
        case e := <- obs_ch:
            obstruction <- ObstructionEvent(e.is_set)
        }
    }
}
