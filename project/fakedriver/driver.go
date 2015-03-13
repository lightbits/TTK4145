package driver

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

func Init(button_pressed chan ButtonEvent,
          floor_reached  chan ReachedFloorEvent,
          stop_pressed   chan StopButtonEvent,
          obstruction    chan ObstructionEvent) {
}
