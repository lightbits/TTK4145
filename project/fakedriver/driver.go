package driver

import (
    "time"
    "fmt"
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

func SetButtonLamp(btn OrderButton, set bool) {
}

func ClearAllButtonLamps() {

}

func SetDoorOpenLamp(on bool) {
}

func OpenDoor() {
}

func CloseDoor() {
}

func SetStopLamp(on bool) {
}

func SetFloorIndicator(floor int) {
}

func Init() {
    fmt.Println("[DRVR]\tInitialized")
}

func Poll(button_pressed chan OrderButton,
          floor_reached  chan int,
          stop_pressed   chan bool,
          obstruction    chan bool) {

    for {
        floor_reached <- 0
        time.Sleep(3 * time.Second)
        button_pressed <- OrderButton{3, ButtonDown}
        button_pressed <- OrderButton{4, ButtonUp}
        time.Sleep(1 * time.Second)
        floor_reached <- 1
        time.Sleep(1 * time.Second)
        floor_reached <- 2
        time.Sleep(1 * time.Second)
        floor_reached <- 3
        time.Sleep(5 * time.Second)
        floor_reached <- 4
        break
    }
}
