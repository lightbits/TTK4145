package main

import (
    "log"
    "encoding/binary"
    "bytes"
)

type order_type int
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
    FromFloor int
    ToFloor   int
    Type      order_type
    TakenBy   lift_id
}

type client_update struct {
    Request    []order // Do we specify ourselves in the TakenBy field?
                       // Maybe that would work. The master would then
                       // see whether or not that is allowed, and update
                       // the list correspondingly.

    Requesting int     // Are we currently requesting anything?
                       // If not, the server will simply acknowledge our
                       // existence.

    // When we take a job that corresponds to some request here,
    // we delete the request from our client status.
}

type master_update struct {
    PendingOrders []order // When we see an order in here with our id on it
                          // we know we should be doing that.
                          // Yeah, that sounds good. Maybe every time we receive
                          // an update, the lift should go through this list, 
                          // and see what its current job is.
}

func PrintOrder(Order order) {
    log.Printf("From:%d\tTo:%d\tType:%d\tTaken: 0x%x\n", 
               Order.FromFloor, Order.ToFloor, 
               Order.Type, Order.TakenBy.IPAddress)
}

func StructMagic() {
    type thing struct {
        ID uint32
        TTL float32
    }

    type message struct {
        Protocol uint32
        Length   uint32
        UserData [512]byte
    }

    // Write to byte array
    Thing := thing{0xdeadbeef, 3.14}
    Buffer := &bytes.Buffer{}
    binary.Write(Buffer, binary.BigEndian, Thing)
    log.Printf("%x\n", Buffer.Bytes())

    // Read back into struct
    Thing = thing{}
    binary.Read(Buffer, binary.BigEndian, &Thing)
    log.Printf("%x %f\n", Thing.ID, Thing.TTL)

    // Pretend that we got a message over the network
    // containing the above information inside UserData
    


    // Message := message{Protocol: 0xabad1dea, Length: 8}
    // copy(Message.UserData[:], []byte("deadbeef40490fda"))

    // // ... got message, time to parse

    // // Thing := thing{}
    // Buffer := bytes.NewBuffer(Message.UserData[:Message.Length])
    // log.Printf("%x\n", Buffer.Bytes())
    // binary.Read(Buffer, binary.BigEndian, &Thing)
    // log.Printf("%x %.2f\n", Thing.ID, Thing.TTL)

}

func main() {
    StructMagic()

    OrderA := order{
        FromFloor: 0,
        ToFloor: 1,
        Type: order_up,
        TakenBy: lift_id{0xabad1dea, 0xbeef},
    }

    OrderB := order{
        FromFloor: 1,
        ToFloor: 2,
        Type: order_down,
        TakenBy: lift_id{0xaabababa, 0xbeef},
    }

    PendingOrders := []order{OrderA, OrderB}
    Update := master_update{PendingOrders}
    for _, Order := range(Update.PendingOrders) {
        PrintOrder(Order)
    }

    log.Println("Hello!")

    /*
    for {
        select {
            case <- UpdateTimer: (goes off every 20 ms or so)
                Send update to master
            case <- ButtonPressed:
                Add a new Request to our client status
            case <- MasterUpdate:
                Synchronize changes
                Go through list of pending orders
                    if PendingOrder[i] is any of ClientStatus.Requests
                        Delete that request

                Give list of jobs to lift controller
                It prioritizes them in a CONSISTENT way, and just works.
                e.g. oldest first, or something actually clever.

                I suppose this means that the lift is a functional thing.
                It constantly reevaluates what it should be doing (going
                up, down, open door?) based on its input. But not quite,
                since the lift has a state as well.
            case <- MasterTimeout:
                Become master.
                Note!
                If we are master, and we receive a MasterUpdate, we need
                to perform an Master Resolution Ritual (MRR).

                If the received MasterUpdate comes from someone with a
                lower IP, we should stop being a master. But first, we
                need to synchronize our pending orders list with them.
                To do that, we enter a Synchronizing waiting state, where
                we accept NO further orders, until we know that the new
                master has synchronized our orders list.

                Do we explicitly ACK? Probably necessary?
                We could become a client, push all orders into our
                Requests field, and wait until the master sends update
                and we see that our requests are correctly received!!

                Alternatively! Do not send client updates to master,
                while finishing _all_ local orders. Hmmm. But what
                about non-local orders...

                Yeah no, we probably need to sync somehow.
        }
    }
    */
}