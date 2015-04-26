package driver
// #include "io.h"
// #cgo CFLAGS: -std=c99
// #cgo LDFLAGS: -L. -lcomedi -lm
import "C"

const NumFloors = 4
const InvalidFloor = -1

const (
    //in port 4
    PORT4               = 3
    OBSTRUCTION         = (0x300+23)
    STOP                = (0x300+22)
    BUTTON_COMMAND1     = (0x300+21)
    BUTTON_COMMAND2     = (0x300+20)
    BUTTON_COMMAND3     = (0x300+19)
    BUTTON_COMMAND4     = (0x300+18)
    BUTTON_UP1          = (0x300+17)
    BUTTON_UP2          = (0x300+16)

    //in port 1
    PORT1               = 2
    BUTTON_DOWN2        = (0x200+0)
    BUTTON_UP3          = (0x200+1)
    BUTTON_DOWN3        = (0x200+2)
    BUTTON_DOWN4        = (0x200+3)
    SENSOR_FLOOR1       = (0x200+4)
    SENSOR_FLOOR2       = (0x200+5)
    SENSOR_FLOOR3       = (0x200+6)
    SENSOR_FLOOR4       = (0x200+7)

    //out port 3
    PORT3               = 3
    MOTORDIR            = (0x300+15)
    LIGHT_STOP          = (0x300+14)
    LIGHT_COMMAND1      = (0x300+13)
    LIGHT_COMMAND2      = (0x300+12)
    LIGHT_COMMAND3      = (0x300+11)
    LIGHT_COMMAND4      = (0x300+10)
    LIGHT_UP1           = (0x300+9)
    LIGHT_UP2           = (0x300+8)

    //out port 2
    PORT2               = 3
    LIGHT_DOWN2         = (0x300+7)
    LIGHT_UP3           = (0x300+6)
    LIGHT_DOWN3         = (0x300+5)
    LIGHT_DOWN4         = (0x300+4)
    LIGHT_DOOR_OPEN     = (0x300+3)
    LIGHT_FLOOR_IND2    = (0x300+1)
    LIGHT_FLOOR_IND1    = (0x300+0)

    //out port 0
    PORT0               = 1
    MOTOR               = (0x100+0)

    //non-existing ports (for alignment)
    BUTTON_DOWN1        = -1
    BUTTON_UP4          = -1
    LIGHT_DOWN1         = -1
    LIGHT_UP4           = -1
)

var up_buttons = [NumFloors]bit_state{
    bit_state{BUTTON_UP1, false},
    bit_state{BUTTON_UP2, false},
    bit_state{BUTTON_UP3, false},
    bit_state{BUTTON_UP4, false}}

var down_buttons = [NumFloors]bit_state{
    bit_state{BUTTON_DOWN1, false},
    bit_state{BUTTON_DOWN2, false},
    bit_state{BUTTON_DOWN3, false},
    bit_state{BUTTON_DOWN4, false}}

var out_buttons = [NumFloors]bit_state{
    bit_state{BUTTON_COMMAND1, false},
    bit_state{BUTTON_COMMAND2, false},
    bit_state{BUTTON_COMMAND3, false},
    bit_state{BUTTON_COMMAND4, false}}

var floor_sensors = [NumFloors]bit_state{
    bit_state{SENSOR_FLOOR1, false},
    bit_state{SENSOR_FLOOR2, false},
    bit_state{SENSOR_FLOOR3, false},
    bit_state{SENSOR_FLOOR4, false}}

var obstructions = [NumFloors]bit_state{
    bit_state{OBSTRUCTION, false}}

var stop_buttons = [NumFloors]bit_state{
    bit_state{STOP, false}}

var up_lights   = []int{LIGHT_UP1,      LIGHT_UP2,      LIGHT_UP3,      LIGHT_UP4}
var down_lights = []int{LIGHT_DOWN1,    LIGHT_DOWN2,    LIGHT_DOWN3,    LIGHT_DOWN4}
var out_lights  = []int{LIGHT_COMMAND1, LIGHT_COMMAND2, LIGHT_COMMAND3, LIGHT_COMMAND4}

func io_init() bool {
    status := C.io_init()
    return status != 0
}

func io_set_bit(channel int) {
    C.io_set_bit(C.int(channel))
}

func io_clear_bit(channel int) {
    C.io_clear_bit(C.int(channel))
}

func io_write_analog(channel, value int) {
    C.io_write_analog(C.int(channel), C.int(value))
}

func io_read_bit(channel int) int {
    return int(C.io_read_bit(C.int(channel)))
}

func io_read_analog(channel int) int {
    return int(C.io_read_analog(C.int(channel)))
}
