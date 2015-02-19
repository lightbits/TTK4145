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

type ReachedFloorEvent struct {
    FloorIndex int
}

type StopButtonEvent struct {
    IsPressed bool
}

type ObstructionEvent struct {
    IsObstructed bool
}

type MotorDirection int
const (
    MotorDirectionUp   = 1
    MotorDirectionStop = 0
    MotorDirectionDown = -1
)

type edge_detect int
const (
    edge_detect_rising = iota
    edge_detect_falling
    edge_detect_both
)

type io_event struct {
    bit    int
    is_set bool
}

func poll(bits [N_FLOORS]bit_state, event chan io_event, edge edge_detect) {
    for {
        for i := 0; i < len(bits); i++ {
            is_set := io_read_bit(bits[i].channel) == 1

            is_rising := !bits[i].was_set && is_set
            is_falling := bits[i].was_set && !is_set

            if ((edge == edge_detect_rising && is_rising) ||
                (edge == edge_detect_falling && is_falling) ||
                (edge == edge_detect_both && (is_rising || is_falling))) {
                event <- io_event{i, is_set}
            }
            bits[i].was_set = is_set
        }

        // If we hammer the io module too much it actually fails to read.
        // So we sleep a little bit.
        time.Sleep(1 * time.Millisecond)
    }
}

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
          floor_reached  chan ReachedFloorEvent,
          stop_pressed   chan StopButtonEvent,
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
        go poll(UP_BUTTONS, up_ch, edge_detect_rising)
        go poll(DOWN_BUTTONS, down_ch, edge_detect_rising)
        go poll(OUT_BUTTONS, out_ch, edge_detect_rising)
    }

    if (floor_reached != nil) {
        go poll(FLOOR_SENSORS, flr_ch, edge_detect_rising)
    }

    if (stop_pressed != nil) {
        go poll(STOP_BUTTONS, stp_ch, edge_detect_rising)
    }

    if (obstruction != nil) {
        go poll(OBSTRUCTIONS, obs_ch, edge_detect_both)
    }

    for {
        select {
        case e := <-up_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonUp}
        case e := <-down_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonDown}
        case e := <-out_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonOut}
        case e := <-flr_ch:
            floor_reached <- ReachedFloorEvent{e.bit}
        case e := <- stp_ch:
            stop_pressed <- StopButtonEvent{e.is_set}
        case e := <- obs_ch:
            obstruction <- ObstructionEvent{e.is_set}
        }
    }
}
