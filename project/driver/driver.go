package driver

import (
    "log"
    "time"
)

type ButtonEvent struct {
    Floor int
    Type  int
}

type FloorEvent       int
type StopEvent        bool
type ObstructionEvent bool

type bit_state struct {
    channel int
    was_set bool
}

type io_event struct {
    bit int
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

        // If we hammer the io module too much it actually fails to read
        // So we sleep a little bit
        time.Sleep(1 * time.Millisecond)
    }
}

func Init(button_pressed chan ButtonEvent,
          floor_reached  chan FloorEvent,
          stop_pressed   chan StopEvent,
          obstruction    chan ObstructionEvent) {

    if (!io_init()) {
        log.Fatal("Failed to initialize driver")
    }

    var up_buttons = [N_FLOORS]bit_state{
        bit_state{BUTTON_UP1, false}, 
        bit_state{BUTTON_UP2, false},
        bit_state{BUTTON_UP3, false},
        bit_state{BUTTON_UP4, false}}

    var down_buttons = [N_FLOORS]bit_state{
        bit_state{BUTTON_DOWN1, false}, 
        bit_state{BUTTON_DOWN2, false},
        bit_state{BUTTON_DOWN3, false},
        bit_state{BUTTON_DOWN4, false}}

    var out_buttons = [N_FLOORS]bit_state{
        bit_state{BUTTON_COMMAND1, false}, 
        bit_state{BUTTON_COMMAND2, false},
        bit_state{BUTTON_COMMAND3, false},
        bit_state{BUTTON_COMMAND4, false}}

    var floor_sensors = [N_FLOORS]bit_state{
        bit_state{SENSOR_FLOOR1, false},
        bit_state{SENSOR_FLOOR2, false},
        bit_state{SENSOR_FLOOR3, false},
        bit_state{SENSOR_FLOOR4, false}}

    var obstructions = [N_FLOORS]bit_state{
        bit_state{OBSTRUCTION, false}}

    var stops = [N_FLOORS]bit_state{
        bit_state{STOP, false}}

    up_ch   := make(chan io_event)
    down_ch := make(chan io_event)
    out_ch  := make(chan io_event)
    flr_ch  := make(chan io_event)
    stp_ch  := make(chan io_event)
    obs_ch  := make(chan io_event)

    if (button_pressed != nil) {
        go poll(up_buttons, up_ch)
        go poll(down_buttons, down_ch)
        go poll(out_buttons, out_ch)
    }

    if (floor_reached != nil) {
        go poll(floor_sensors, flr_ch)
    }

    if (stop_pressed != nil) {
        go poll(stops, stp_ch)
    }

    if (obstruction != nil) {
        go poll(obstructions, obs_ch)
    }

    for {
        select {
        case e := <-up_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, 0}
            }
        case e := <-down_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, 1}
            }
        case e := <-out_ch:
            if (e.is_set) {
                button_pressed <- ButtonEvent{e.bit, 2}
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