/*
TODO:
* Split up into modules: queue, client, master
* Add timer that fires when a client has not performed his order in a while
* Make client and master more clean
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
    Timer           *time.Timer
    HasTimedOut     bool
}

type ClientData struct {
    LastPassedFloor int
    Requests        []Order
}

type MasterData struct {
    AssignedBackup network.ID
    Orders         []Order
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
    if err == nil {
        for _, o := range(result.Orders) {
            if o.TakenBy == network.InvalidID {
                log.Fatal("[CLIENT]\tA non-taken order was received")
            }
        }
    }

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

func DistanceSqrd(a, b int) int {
    return (a - b) * (a - b)
}

func ClosestActiveLift(clients map[network.ID]Client, floor int) network.ID {
    closest_df := 100
    closest_id := network.InvalidID
    for id, client := range(clients) {
        if client.HasTimedOut {
            continue
        }
        df := DistanceSqrd(client.LastPassedFloor, floor)
        if df < closest_df {
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

/*
This is not a very good prioritization algorithm, but
we have the data we need if we want to make it better.
--
The distribution/prioritization works in two steps.
One is a global pass, which distributes all non-taken
orders across all the lifts, based purely on proximity.

The second pass works on each individual lift, picking
out a single order that should be prioritized. If the
lift is idle (i.e. it has reached its target), the next
order is chosen to be whichever is closest.

If the lift is moving, we check if there is an order
for the same direction that is closer along its path.
If so, we make that the priority.

Note that if the lift completes an order, the order will
be deleted from the master side. When this happens, the
lift might not have a target floor. But this is OK, since
we interpret this as the lift being idle.
*/
func DistributeWork(clients map[network.ID]Client, orders []Order) {
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

    for id, c := range(clients) {
        target_floor := -1
        current_pri  := -1
        for index, order := range(orders) {
            if order.TakenBy == id && order.Priority {
                target_floor = order.Button.Floor
                current_pri = index
            }
        }

        better_pri := -1
        if target_floor >= 0 {
            better_pri = ClosestOrderAlong(id, orders, c.LastPassedFloor, target_floor)
        } else {
            better_pri = ClosestOrderNear(id, orders, c.LastPassedFloor)
        }

        if better_pri >= 0 {
            if current_pri >= 0 {
                orders[current_pri].Priority = false
            }
            orders[better_pri].Priority = true
        }
    }
}

func IsSameOrder(a, b Order) bool {
    return a.Button.Floor == b.Button.Floor &&
           a.Button.Type  == b.Button.Type
}

func MasterLoop(c Channels, backup network.ID, initial_queue []Order) {
    TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL    := 250 * time.Millisecond

    time_to_send     := time.NewTicker(SEND_INTERVAL)
    client_timed_out := make(chan network.ID)

    orders  := initial_queue
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
            // TODO: Seperate this into a different module? Network module?
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
                for i, o := range(orders) {
                    if IsSameOrder(o, r) {
                        if r.Done {
                            o.Done = true
                            orders[i] = o
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
            for i := 0; i < len(orders); i++ {
                if orders[i].Done {
                    orders = append(orders[:i], orders[i+1:]...)
                    i--
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

func RemoveAcknowledgedRequests(requests, orders []Order) []Order {
    for i := 0; i < len(requests); i++ {
        r := requests[i]
        master_has_it := false
        acknowledged := false
        for _, o := range(orders) {
            if IsSameOrder(r, o) {
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

func TestRemoveAcknowledgedRequests() bool {
    order := Order {
        Button: driver.OrderButton {
            Floor: 5,
            Type: driver.ButtonUp,
        },
    }
    orders := make([]Order, 0)
    orders = append(orders, order)
    order.Done = false
    order.Button.Floor = 5
    requests := make([]Order, 0)
    requests = append(requests, order)
    requests = RemoveAcknowledgedRequests(requests, orders)
    fmt.Println(requests)
    return len(requests) == 0
}

func ClientLoop(c Channels, master network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    orders := make([]Order, 0) // Local copy of master's queue
    requests := make([]Order, 0) // Unacknowledged local events

    our_id := network.GetMachineID()
    last_passed_floor := 0
    is_backup := false

    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
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
            order := Order {
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

            driver.ClearAllButtonLamps()

            orders = data.Orders
            for _, o := range(orders) {
                if !(o.Button.Type == driver.ButtonOut && o.TakenBy != our_id) {
                    driver.SetButtonLamp(o.Button, true)
                }
                if o.TakenBy == our_id && o.Priority {
                    is_done := false
                    for _, r := range(requests) {
                        if IsSameOrder(o, r) && r.Done {
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
            driver.ClearAllButtonLamps()
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

    // TestDriver(channels)

    // fmt.Println(TestRemoveAcknowledgedRequests())

    if start_as_master {
        initial_queue := make([]Order, 0)
        go WaitForBackup(channels, initial_queue)
    }

    go network.ClientWorker(channels.from_master, channels.to_master)
    WaitForMaster(channels, nil)
}
