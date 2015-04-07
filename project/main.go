package main

import (
    "time"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "./lift"
    "./network"
    // "./driver"
    "./fakedriver"
)

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type Client struct {
    ID              network.ID
    LastPassedFloor int
    TargetFloor     int
    Timer           *time.Timer
    HasTimedOut     bool
}

type ClientData struct {
    LastPassedFloor int
    TargetFloor     int
    Requests        []Order
}

type MasterData struct {
    AssignedBackup network.ID
    Orders         []Order
}

type Channels struct {
    completed_floor chan bool
    reached_target  chan bool
    button_pressed  chan driver.OrderButton
    floor_reached   chan int
    stop_button     chan bool
    obstruction     chan bool
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

func WaitForBackup(c Channels, initial_queue []Order) {
    go network.MasterWorker(c.from_client, c.to_clients)
    machine_id := network.GetMachineID()
    fmt.Println("[MASTER]\tRunning on machine", machine_id)
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.from_client:
            // DEBUG:
            MasterLoop(c, packet.Address)

            // if packet.Address != machine_id {
            //     MasterLoop(c, packet.Address)
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

func DistanceSqrd(a, b int) int {
    return (a - b) * (a - b)
}

func ClosestActiveLift(clients map[network.ID]Client, floor int) network.ID {
    closest_df := -1
    closest_id := network.InvalidID
    for id, client := range(clients) {
        if client.HasTimedOut {
            continue
        }
        df := DistanceSqrd(client.LastPassedFloor, floor)
        if closest_df == -1 || df < closest_df {
            closest_df = df
            closest_id = id
        }
    }
    return closest_id
}

func ClosestOrderNear(owner network.ID, orders []Order, floor int) int {
    closest_i := -1
    closest_d := -1
    for i, o := range(orders) {
        if o.TakenBy != owner {
            continue
        }
        d := DistanceSqrd(o.Button.Floor, floor)
        if closest_i == -1 || d < closest_d {
            closest_i = i
            closest_d = d
        }
    }
    return closest_i
}

func ClosestOrderAlong(owner network.ID, orders []Order, from, to int) int {
    closest_i := -1
    closest_d := -1
    for i, o := range(orders) {
        if o.TakenBy != owner {
            continue
        }
        // Deliberately not using o.Floor >= from, since
        // the lift might not actually be at its last passed
        // floor by the time we distribute work.
        in_range   := o.Button.Floor > from && o.Button.Floor <= to
        dir_up     := to - from > 0 // Likewise, these are not using = since we
        dir_down   := to - from < 0 // assert that LPF != TF when calling this
        order_up   := o.Button.Type == driver.ButtonUp
        order_down := o.Button.Type == driver.ButtonDown
        order_out  := o.Button.Type == driver.ButtonOut
        if in_range && ((dir_up   && (order_up   || order_out)) ||
                        (dir_down && (order_down || order_out))) {
            d := DistanceSqrd(o.Button.Floor, from)
            if closest_i == -1 || d < closest_d {
                closest_i = i
                closest_d = d
            }
        }
    }
    return closest_i
}

func DistributeWork(clients map[network.ID]Client, orders []Order) {
    // Broad-phase distribution
    for i, o := range(orders) {
        if (o.Button.Type != driver.ButtonOut) &&
           (o.TakenBy == network.InvalidID ||
            clients[o.TakenBy].HasTimedOut) {

            closest := ClosestActiveLift(clients, o.Button.Floor)
            if closest == network.InvalidID {
                log.Fatal("Cannot distribute work when there are no lifts!")
            }
            o.TakenBy = closest
            orders[i] = o
        }
    }

    // Narrow-phase distribution (sort each lift queue)
    for id, c := range(clients) {
        // If the client is already heading towards a floor, we
        // don't want to change its direction. But we if there
        // is a new floor that is closer along the way, we can
        // stop there first. But only if that order is also
        // headed the same way...

        // Note that the LPF will eventually equal TF, as the client
        // can only go to the one floor which master marks as PRIORITY.
        if c.LastPassedFloor == c.TargetFloor {
            closest := ClosestOrderNear(id, orders, c.LastPassedFloor)
            orders[closest].Priority = true
        } else {
            closest := ClosestOrderAlong(id, orders, c.LastPassedFloor, c.TargetFloor)
            orders[closest].Priority = true
        }
    }
}

func IsSameOrder(a, b Order) bool {
    return a.Button.Floor == b.Button.Floor &&
           a.Button.Type  == b.Button.Type
}

func MasterLoop(c Channels, backup network.ID) {
    SEND_INTERVAL    := 2 * time.Second
    TIMEOUT_INTERVAL := 5 * time.Second

    time_to_send     := time.NewTicker(SEND_INTERVAL)
    client_timed_out := make(chan network.ID)

    orders  := make([]Order, 0)
    clients := make(map[network.ID]Client)

    fmt.Println("[MASTER]\tStarting master with backup", backup)
    for {
        select {
        case packet := <- c.from_client:
            fmt.Println("[MASTER]\tClient said", string(packet.Data))

            data, err := DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }

            // Add client if new, and update information
            sender_id := packet.Address
            client, exists := clients[sender_id]
            if exists {
                client.Timer.Reset(TIMEOUT_INTERVAL)
                client.LastPassedFloor = data.LastPassedFloor
                clients[sender_id] = client
            } else {
                timer := time.NewTimer(TIMEOUT_INTERVAL)
                client := Client {
                    ID:              sender_id,
                    LastPassedFloor: data.LastPassedFloor,
                    TargetFloor:     data.TargetFloor,
                    Timer:           timer,
                }
                clients[sender_id] = client
                go ListenForClientTimeout(sender_id, timer, client_timed_out)
            }

            // Synchronize our list of jobs with any new or finished
            // jobs given by the client's requests
            requests := data.Requests
            for _, r := range(requests) {

                is_new_order := true
                for _, o := range(orders) {
                    if IsSameOrder(o, r) {
                        if r.Done {
                            o.Done = true
                        }
                        is_new_order = false
                    }
                }

                if r.Button.Type == driver.ButtonOut {
                    r.TakenBy = sender_id
                }

                if is_new_order {
                    orders = append(orders, r)
                }
            }

            // Delete finished jobs
            for i, o := range(orders) {
                if o.Done {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }

        case <- time_to_send.C:
            DistributeWork(clients, orders)
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
                client.HasTimedOut = true
                clients[who] = client
            } else {
                log.Fatal("[MASTER]\tA non-existent client timed out")
            }
        }
    }
}

func WaitForMaster(c Channels, remaining_orders []Order) {
    fmt.Println("[CLIENT]\tWaiting for master...")
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
        case <- c.button_pressed:
        case <- c.floor_reached:
        case <- c.stop_button: // ignore
        case <- c.obstruction: // ignore
        }

    }
}

func ClientLoop(c Channels, master network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 2 * time.Second

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    orders   := make([]Order, 0) // Local copy of master's queue
    requests := make([]Order, 0)

    our_id := network.GetMachineID()

    target_floor      := 0
    last_passed_floor := 0

    requests = append(requests, Order {
        Button: driver.OrderButton{5, driver.ButtonUp}})

    is_backup := false
    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case packet := <- c.from_master:
            fmt.Println("[CLIENT]\tMaster said", string(packet.Data))
            master_timeout.Reset(MASTER_TIMEOUT_INTERVAL)
            data, err := DecodeMasterPacket(packet.Data)
            if err != nil {
                break
            }

            if data.AssignedBackup == our_id {
                fmt.Println("[CLIENT]\tWe are the backup!")
                is_backup = true
            } else {
                is_backup = false
            }

            // TODO: Clear all button lamps

            orders = data.Orders
            for _, o := range(orders) {

                if o.TakenBy == network.InvalidID {
                    log.Fatal("[CLIENT]\tA non-taken order was received")
                }

                // driver.SetButtonLamp(o.Button, true)

                if o.TakenBy == our_id && o.Priority {
                    target_floor = o.Button.Floor
                    fmt.Println("[CLIENT]\tTarget floor:", target_floor)
                }
            }

            // Clear requests that are acknowledged
            for i, r := range(requests) {
                found := false
                safe_to_delete := false
                for _, o := range(orders) {
                    if IsSameOrder(r, o) {
                        found = true
                        if r.Done && !o.Done {
                            // We have finished the order, but are waiting
                            // for the master to acknowledge that
                        } else if r.Done && o.Done {
                            // This shouldn't happen?
                        } else if !r.Done && o.Done {
                            // This shouldn't happen either?
                        } else if !r.Done && !o.Done {
                            // Aha! Now it is safe to remove it from
                            // out requests, as the master has acked it.
                            safe_to_delete = true
                        }
                    }
                }
                if !found && r.Done {
                    // This means the master has acknowledged that the
                    // order was finished, and it is safe to delete
                    // the request.
                    safe_to_delete = true
                }

                if safe_to_delete {
                    requests = append(requests[:i], requests[i+1:]...)
                }
            }

        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
                // Note that if there were any unacknowledged new orders
                // or finished orders in requests, they will be left out.
                // If they were new orders, it is ok since we have yet to
                // give user feedback. If they were completed orders...
                // maybe ok?
                WaitForBackup(c, orders)
            }

        case <- time_to_send.C:
            data := ClientData {
                LastPassedFloor: last_passed_floor,
                TargetFloor:     target_floor,
                Requests:        requests,
            }
            c.to_master <- network.Packet {
                Data: EncodeClientData(data),
            }

        // case button := <- c.button_pressed:

        // case floor := <- c.floor_reached:
        // case stopped := <- c.stop_button:
        // case obstructed := <- c.obstruction:
        // case <- c.completed_floor:
        }
    }
}

func TestNetwork(channels Channels) {
    go network.ClientWorker(channels.from_master, channels.to_master)
    go network.MasterWorker(channels.from_client, channels.to_clients)
    t1 := time.NewTimer(1*time.Second)
    t2 := time.NewTimer(2*time.Second)
    for {
        select {
        case <- t1.C:
            fmt.Println("[NETTEST]\tSending to all clients")
            channels.to_clients <- network.Packet{
                Data: []byte("A")}

        case <- t2.C:
            fmt.Println("[NETTEST]\tSending to any master")
            channels.to_master <- network.Packet{
                Data: []byte("BB")}

        case p := <- channels.from_client:
            fmt.Println("[NETTEST]\tClient sent:", len(p.Data), "bytes from", p.Address)

        case p := <- channels.from_master:
            fmt.Println("[NETTEST]\tMaster sent:", len(p.Data), "bytes from", p.Address)
        }
    }
}

func TestDriver(channels Channels) {
    for {
        select {
        case btn := <- channels.button_pressed:
            fmt.Println("[TEST]\tButton pressed")
            driver.SetButtonLamp(btn, true)
        case <- channels.floor_reached:
            fmt.Println("[TEST]\tFloor reached")
        case <- channels.stop_button:
            fmt.Println("[TEST]\tStop button pressed")
        case <- channels.obstruction:
            fmt.Println("[TEST]\tObstruction changed")
        }
    }
}

func main() {
    var start_as_master bool
    flag.BoolVar(&start_as_master, "master", false, "Start as master")
    flag.Parse()

    var channels Channels
    channels.completed_floor = make(chan bool)
    channels.reached_target  = make(chan bool)
    channels.button_pressed  = make(chan driver.OrderButton)
    channels.floor_reached   = make(chan int)
    channels.stop_button     = make(chan bool)
    channels.obstruction     = make(chan bool)
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
        channels.completed_floor,
        channels.reached_target)

    // TestDriver(channels)

    if start_as_master {
        go WaitForBackup(channels, nil)
    }

    go network.ClientWorker(channels.from_master, channels.to_master)
    WaitForMaster(channels, nil)
}
