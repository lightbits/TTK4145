package driver

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
}
