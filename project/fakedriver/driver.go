package driver

import (
    // "time"
    "../logger"
    "bufio"
    "os"
    "fmt"
)

const N_FLOORS = 4
const INVALID_FLOOR = -1

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

func listenForUserInput(input chan string) {
    reader := bufio.NewReader(os.Stdin)
    for {
        // fmt.Printf(">>")
        line, _, err := reader.ReadLine()
        if err != nil {
            fmt.Println(err)
        }
        input <- string(line)
    }
}

func MotorUp() {
    println(logger.Info, "Go up")
}

func MotorDown() {
    println(logger.Info, "Go down")
}

func MotorStop() {
    println(logger.Info, "Stop")
}

func SetButtonLamp(btn OrderButton, set bool) {
}

func ClearAllButtonLamps() {

}

func SetDoorOpenLamp(on bool) {
}

func OpenDoor() {
    println(logger.Info, "Open door")
}

func CloseDoor() {
    println(logger.Info, "Close door")
}

func SetStopLamp(on bool) {
}

func SetFloorIndicator(floor int) {
}

func Init() {
    println(logger.Info, "Initialized")
}

func Poll(button_pressed chan OrderButton,
          floor_reached  chan int,
          stop_pressed   chan bool,
          obstruction    chan bool) {

    input := make(chan string)
    go listenForUserInput(input)
    for {
        line := <- input
        var what string
        var arg int
        fmt.Sscanf(line, "%s%d", &what, &arg)
        if what == "f" {
            floor_reached <- arg
        } else if what == "u" {
            button_pressed <- OrderButton{arg, ButtonUp}
        } else if what == "d" {
            button_pressed <- OrderButton{arg, ButtonDown}
        } else if what == "o" {
            button_pressed <- OrderButton{arg, ButtonOut}
        }
    }
    // for {
    //     floor_reached <- 0
    //     time.Sleep(3 * time.Second)
    //     button_pressed <- OrderButton{3, ButtonDown}
    //     button_pressed <- OrderButton{4, ButtonUp}
    //     button_pressed <- OrderButton{4, ButtonOut}
    //     time.Sleep(1 * time.Second)
    //     floor_reached <- 1
    //     time.Sleep(1 * time.Second)
    //     floor_reached <- 2
    //     time.Sleep(1 * time.Second)
    //     floor_reached <- 3
    //     time.Sleep(5 * time.Second)
    //     floor_reached <- 4
    //     break
    // }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "DRIVER", args...)
}