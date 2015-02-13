package main

import (
    "log"
    "encoding/binary"
    "bytes"
    "./driver"
)

type order_type int32
const (
    order_up   = 1
    order_out  = 0
    order_down = -1
)

type lift_id struct {
    IPAddress uint32
    Port      uint16
}

type order struct {
    Floor     int
    Type      order_type
    TakenBy   lift_id
    Finished  bool
}

type client_update struct {
    Requests   []order
    Requesting int
}

type master_update struct {
    PendingOrders []order
}

func PrintOrder(Order order) {
    log.Printf("Floor: %d\tType:%d\tTaken: 0x%x\n", 
               Order.Floor, Order.Type, Order.TakenBy.IPAddress)
}

func NetworkMockup() {
    type network_message struct {
        Protocol     uint32
        Length       uint32
        UserData     []byte
        // Checksum
    }

    type request struct {
        Floor   int32
        Type    order_type
        TakenBy lift_id
    }

    // Let's send a request over the network!
    r := request{Floor: 3, Type: order_up, TakenBy: lift_id{0x81F1BB88, 20012}}

    // First! We need to wrap it into a network_message packet
    // so that our network module can transmit it.

    // Convert data to byte array
    b := &bytes.Buffer{}
    binary.Write(b, binary.BigEndian, r)

    // Wrap into packet
    const CL_STATUS_PROTOCOL = 0xdeadbeef
    packet := network_message{Protocol: CL_STATUS_PROTOCOL, 
                              Length: uint32(b.Len()), 
                              UserData: b.Bytes()}

    // Send over network....
    // ...

    // Receiver side, we need to parse the packet
    if (packet.Protocol != CL_STATUS_PROTOCOL) {
        log.Println("Received an invalid packet")
        return
    }

    // And read into the struct type
    b = bytes.NewBuffer(packet.UserData[:packet.Length])
    r = request{}
    binary.Read(b, binary.BigEndian, &r)

    log.Println("Received a client status update")
    log.Printf("Floor: %d Type: %d TakenBy: %d.%d.%d.%d\n",
               r.Floor, r.Type, 
               (r.TakenBy.IPAddress >> 24) & 0xff,
               (r.TakenBy.IPAddress >> 16) & 0xff,
               (r.TakenBy.IPAddress >> 8) & 0xff,
               r.TakenBy.IPAddress & 0xff)
}

func StructMagic() {
    type thing struct {
        ID  uint32
        TTL float32
    }

    type message struct {
        Protocol uint32
        Length   uint32
        UserData [512]byte
    }

    // Write to byte array
    t := thing{0xdeadbeef, 3.141593}
    b := &bytes.Buffer{}
    binary.Write(b, binary.BigEndian, t)
    log.Printf("%x\n", b.Bytes())

    // Pretend that we got a message over network
    m := message{Protocol: 0xabad1dea, Length: 8}
    copy(m.UserData[:m.Length], b.Bytes())
    log.Printf("%x\n", m.UserData[:m.Length])

    // Parse message and convert into a readable thing
    b = bytes.NewBuffer(m.UserData[:m.Length])
    t = thing{}
    binary.Read(b, binary.BigEndian, &t)
    log.Printf("%x %f\n", t.ID, t.TTL)

    // Read back into struct
    // Thing = thing{}
    // binary.Read(Buffer, binary.BigEndian, &Thing)
    // log.Printf("%x %f\n", Thing.ID, Thing.TTL)

}

func main() {
    // StructMagic()
    NetworkMockup()

    button_pressed  := make(chan driver.ButtonEvent)
    floor_reached   := make(chan driver.FloorEvent)
    stop_pressed    := make(chan driver.StopEvent)
    obstruction     := make(chan driver.ObstructionEvent)

    go driver.Init(button_pressed, floor_reached, stop_pressed, obstruction)

    for {
        select {
        case button := <- button_pressed:
            log.Println("Button ", button)
        case obstructed := <- obstruction:
            log.Println("Obstructed ", obstructed)
        case <- stop_pressed:
            log.Println("Stop")
        case floor := <- floor_reached:
            log.Println("Floor ", floor)
        }
    }

    OrderA := order{
        Floor: 0,
        Type: order_up,
        TakenBy: lift_id{0xabad1dea, 0xbeef},
    }

    OrderB := order{
        Floor: 1,
        Type: order_down,
        TakenBy: lift_id{0xaabababa, 0xbeef},
    }

    PendingOrders := []order{OrderA, OrderB}
    Update := master_update{PendingOrders}
    for _, Order := range(Update.PendingOrders) {
        PrintOrder(Order)
    }

    log.Println("Hello!")
}