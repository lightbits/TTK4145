/*
TODO:
* Split up into modules: queue, client, master
* Add timer that fires when a client has not performed his order in a while
* Make client and master more clean
* Implement WaitForMaster
*/

package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "./lift"
    "./network"
    "./queue"
    "./driver"
)

type ClientData struct {
    LastPassedFloor int
    Requests        []queue.Order
}

type MasterData struct {
    AssignedBackup network.ID
    Orders         []queue.Order
}

type Channels struct {
    // Lift events
    last_passed_floor_changed chan int
    target_floor_changed      chan int
    completed_floor           chan int

    // Driver events
    button_pressed  chan driver.OrderButton
    floor_reached   chan int
    stop_button     chan bool
    obstruction     chan bool

    // Network events
    to_master       chan network.Packet
    to_clients      chan network.Packet
    from_master     chan network.Packet
    from_client     chan network.Packet
}

func DecodeMasterPacket(b []byte) (MasterData, error) {
    var result MasterData
    err := json.Unmarshal(b, &result)
    return result, err
}

func DecodeClientPacket(b []byte) (ClientData, error) {
    var result ClientData
    err := json.Unmarshal(b, &result)
    return result, err
}

func EncodeMasterData(m MasterData) []byte {
    result, err := json.Marshal(m)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func EncodeClientData(c ClientData) []byte {
    result, err := json.Marshal(c)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func WaitForBackup(c Channels, initial_queue []queue.Order,) {

    machine_id := network.GetMachineID()
    fmt.Println("[MASTER]\tRunning on machine", machine_id)
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.from_client:
            // DEBUG:
            MasterLoop(c, packet.Address, initial_queue)

            // if packet.Address != machine_id {
            //     MasterLoop(c, packet.Address, initial_queue)
            //     return
            // } else {
            //     fmt.Println("[MASTER]\tCannot use own machine as backup client")
            // }
        }
    }
}

func ListenForClientTimeout(id network.ID, timer *time.Timer, timeout chan network.ID) {
    select {
    case <- timer.C:
        timeout <- id
    }
}

func AddNewOrders(requests, orders []queue.Order, sender network.ID) []queue.Order {
    for _, r := range(requests) {
        if r.Button.Type == driver.ButtonOut {
            r.TakenBy = sender
        }

        if queue.IsNewOrder(r, orders) {
            orders = append(orders, r)
        }
    }
    return orders
}

func DeleteDoneOrders(requests, orders []queue.Order) []queue.Order {
    for i := 0; i < len(orders); i++ {
        for _, r := range(requests) {
            if queue.IsSameOrder(orders[i], r) && r.Done {
                orders[i].Done = true
            }
        }
        if orders[i].Done {
            orders = append(orders[:i], orders[i+1:]...)
            i--
        }
    }
    return orders
}

func MasterLoop(c Channels, backup network.ID, initial_queue []queue.Order) {

    TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL    := 250 * time.Millisecond

    time_to_send     := time.NewTicker(SEND_INTERVAL)
    client_timed_out := make(chan network.ID)

    orders := initial_queue
    if initial_queue == nil {
        orders = make([]queue.Order, 0)
    }

    clients := make(map[network.ID]queue.Client)

    fmt.Println("[MASTER]\tStarting master with backup", backup)
    for {
        select {
        case packet := <- c.from_client:

            data, err := DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }
            fmt.Println("[MASTER]\tClient said", data)

            sender_id := packet.Address
            client, exists := clients[sender_id]
            if !exists {
                fmt.Println("[MASTER]\tAdding new client", sender_id)
                timer := time.NewTimer(TIMEOUT_INTERVAL)
                client = queue.Client {
                    ID: sender_id,
                    Timer: timer,
                }
                go ListenForClientTimeout(sender_id, timer, client_timed_out)
            }
            client.Timer.Reset(TIMEOUT_INTERVAL)
            client.HasTimedOut = false
            client.LastPassedFloor = data.LastPassedFloor
            clients[sender_id] = client

            orders = AddNewOrders(data.Requests, orders, sender_id)
            orders = DeleteDoneOrders(data.Requests, orders)


        case <- time_to_send.C:
            queue.DistributeWork(clients, orders)
            data := MasterData {
                AssignedBackup: backup,
                Orders:         orders,
            }
            c.to_clients <- network.Packet {
                Data: EncodeMasterData(data),
            }

        case who := <- client_timed_out:
            fmt.Println("[MASTER]\tClient", who, "timed out")
            client, exists := clients[who]
            if exists {
                queue.RemoveExternalAssignments(orders, who)
                client.HasTimedOut = true
                clients[who] = client
            }
            if who == backup {
                WaitForBackup(c, orders)
            }
        }
    }
}

func WaitForMaster(c Channels, remaining_orders []queue.Order) {
    fmt.Println("[CLIENT]\tWaiting for ..")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tHeard a master!")
            ClientLoop(c, packet.Address)
            return

        case <- time_to_ping.C:
            c.to_master <- network.Packet {
                Data: []byte("Ping"),
            }
        case button := <- c.button_pressed:
            if button.Type == driver.ButtonOut {
                // TODO: Add to order list
            }
        case <- c.floor_reached:
        case <- c.stop_button: // ignore
        case <- c.obstruction: // ignore
        }

    }
}

func RemoveAcknowledgedRequests(requests, orders []queue.Order) []queue.Order {
    for i := 0; i < len(requests); i++ {
        r := requests[i]
        master_has_it := false
        acknowledged := false
        for _, o := range(orders) {
            if queue.IsSameOrder(r, o) {
                master_has_it = true
                if r.Done == o.Done {
                    acknowledged = true
                }
            }
        }
        if !master_has_it && r.Done {
            acknowledged = true
        }
        if acknowledged {
            requests = append(requests[:i], requests[i+1:]...)
            i--
        }
    }
    return requests
}

func ClientLoop(c Channels, master network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    orders := make([]queue.Order, 0) // Local copy of master's queue
    requests := make([]queue.Order, 0) // Unacknowledged local events

    our_id := network.GetMachineID()
    last_passed_floor := 0
    is_backup := false

    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
                go network.MasterWorker(c.from_client, c.to_clients)
                go WaitForBackup(c, orders)
            }

        case <- time_to_send.C:
            data := ClientData {
                LastPassedFloor: last_passed_floor,
                Requests:        requests,
            }
            c.to_master <- network.Packet {
                Data: EncodeClientData(data),
            }

        case button := <- c.button_pressed:
            fmt.Println("[CLIENT]\tA button was pressed")
            order := queue.Order {
                Button: button,
            }
            requests = append(requests, order)

        case floor := <- c.last_passed_floor_changed:
            last_passed_floor = floor

        case floor := <- c.completed_floor:
            for _, o := range(orders) {
                if o.TakenBy == our_id && o.Button.Floor == floor {
                    o.Done = true
                    requests = append(requests, o)
                }
            }

        case packet := <- c.from_master:
            master_timeout.Reset(MASTER_TIMEOUT_INTERVAL)
            data, err := DecodeMasterPacket(packet.Data)
            if err != nil {
                break
            }
            fmt.Println("[CLIENT]\tMaster said", data)

            if data.AssignedBackup == our_id {
                is_backup = true
            } else {
                is_backup = false
            }

            driver.ClearAllButtonLamps()

            orders = data.Orders
            for _, o := range(orders) {
                if !(o.Button.Type == driver.ButtonOut && o.TakenBy != our_id) {
                    driver.SetButtonLamp(o.Button, true)
                }
                if o.TakenBy == our_id && o.Priority {
                    is_done := false
                    for _, r := range(requests) {
                        if queue.IsSameOrder(o, r) && r.Done {
                            is_done = true
                        }
                    }
                    if !is_done {
                        c.target_floor_changed <- o.Button.Floor
                    }
                    fmt.Println("[CLIENT]\tTarget floor:", o.Button.Floor)
                }
            }

            requests = RemoveAcknowledgedRequests(requests, orders)
        }
    }
}

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels Channels

    // Lift events
    channels.last_passed_floor_changed = make(chan int)
    channels.target_floor_changed      = make(chan int)
    channels.completed_floor           = make(chan int)

    // Driver events
    channels.button_pressed  = make(chan driver.OrderButton)
    channels.floor_reached   = make(chan int)
    channels.stop_button     = make(chan bool)
    channels.obstruction     = make(chan bool)

    // Network events
    channels.to_master       = make(chan network.Packet)
    channels.to_clients      = make(chan network.Packet)
    channels.from_master     = make(chan network.Packet)
    channels.from_client     = make(chan network.Packet)

    driver.Init()

    go driver.Poll(
        channels.button_pressed,
        channels.floor_reached,
        channels.stop_button,
        channels.obstruction)

    go lift.Init(
        channels.floor_reached,
        channels.last_passed_floor_changed,
        channels.target_floor_changed,
        channels.completed_floor,
        channels.stop_button,
        channels.obstruction)

    if start_as_master {
        go network.MasterWorker(channels.from_client, channels.to_clients)
        go WaitForBackup(channels, nil)
    }

    go network.ClientWorker(channels.from_master, channels.to_master)
    WaitForMaster(channels, nil)
}
