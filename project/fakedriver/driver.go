package driver

import (
    "time"
)

type ButtonType int
const (
    ButtonUp ButtonType = iota
    ButtonDown
    ButtonOut
)

type OrderButton struct {
    Floor int
    Type  ButtonType
}

func MotorUp() {
}

func MotorDown() {
}

func MotorStop() {
}

func SetButtonLamp(btn ButtonType, floor int, set bool) {
}

func SetDoorOpenLamp(on bool) {
}

func SetStopLamp(on bool) {
}

func SetFloorIndicator(floor int) {
}

func Init(button_pressed chan OrderButton,
          floor_reached  chan int,
          stop_pressed   chan bool,
          obstruction    chan bool) {

    for {
        button_pressed <- OrderButton{3, ButtonDown}
        button_pressed <- OrderButton{4, ButtonUp}
        floor_reached <- 2
        time.Sleep(1 * time.Second)
    }
}
