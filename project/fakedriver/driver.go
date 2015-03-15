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
    fmt.Println("[DRVR]\tMotor up")
}

func MotorDown() {
    fmt.Println("[DRVR]\tMotor down")
}

func MotorStop() {
    fmt.Println("[DRVR]\tMotor stop")
}

func SetButtonLamp(btn OrderButton, set bool) {
    if set {
        fmt.Println("[DRVR]\tLit button", btn.Type, "at", btn.Floor)
    } else {
        fmt.Println("[DRVR]\tUnlit button", btn.Type, "at", btn.Floor)
    }
}

func SetDoorOpenLamp(on bool) {
    fmt.Println("[DRVR]\tDoor open =", on)
}

func SetStopLamp(on bool) {
    fmt.Println("[DRVR]\tStop lamp on =", on)
}

func SetFloorIndicator(floor int) {
    fmt.Println("[DRVR]\tFloor indicator =", floor)
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
        time.Sleep(1 * time.Second)
        button_pressed <- OrderButton{3, ButtonDown}
        button_pressed <- OrderButton{4, ButtonUp}
        time.Sleep(1 * time.Second)
        floor_reached <- 1
        time.Sleep(1 * time.Second)
        floor_reached <- 2
        time.Sleep(1 * time.Second)
        floor_reached <- 3
        time.Sleep(4 * time.Second)
        floor_reached <- 4
    }
}
