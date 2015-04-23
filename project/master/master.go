package master

import (
    "../queue"
    "../com"
    "../network"
    "../driver"
    "fmt"
    "time"
)

func WaitForBackup(c               com.Channels,
                   initial_queue   []com.Order,
                   initial_clients map[network.ID]com.Client) {

    machine_id := network.GetMachineID()
    fmt.Println("[MASTER]\tRunning on machine", machine_id)
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.FromClient:
            // DEBUG:
            // MasterLoop(c, packet.Address, initial_queue, initial_clients)

            if packet.Address != machine_id {
                MasterLoop(c, packet.Address, initial_queue, initial_clients)
                return
            } else {
                fmt.Println("[MASTER]\tCannot use own machine as backup client")
            }
        }
    }
}

func ListenForClientTimeout(id network.ID, timer *time.Timer, timeout chan network.ID) {
    select {
    case <- timer.C:
        timeout <- id
    }
}

func AddNewOrders(requests, orders []com.Order, sender network.ID) []com.Order {
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

func DeleteDoneOrders(requests, orders []com.Order) []com.Order {
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

func RemoveExternalAssignments(orders []com.Order, who network.ID) {
    for i, o := range(orders) {
        if o.TakenBy == who && o.Button.Type != driver.ButtonOut {
            o.TakenBy = network.InvalidID
            orders[i] = o
        }
    }
}

func MasterLoop(c               com.Channels,
                backup          network.ID,
                initial_queue   []com.Order,
                initial_clients map[network.ID]com.Client) {

    TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL    := 250 * time.Millisecond

    time_to_send     := time.NewTicker(SEND_INTERVAL)
    client_timed_out := make(chan network.ID)

    orders := initial_queue
    if initial_queue == nil {
        orders = make([]com.Order, 0)
    }

    clients := make(map[network.ID]com.Client)
    if initial_clients != nil {
        for _, c := range(initial_clients) {
            c.AliveTimer = time.NewTimer(TIMEOUT_INTERVAL)
            clients[c.ID] = c
            go ListenForClientTimeout(c.ID, c.AliveTimer, client_timed_out)
        }
    }

    fmt.Println("[MASTER]\tStarting master with backup", backup)
    for {
        select {
        case packet := <- c.FromClient:

            data, err := com.DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }
            fmt.Println("[MASTER]\tClient said", data)

            sender_id := packet.Address
            client, exists := clients[sender_id]
            if !exists {
                fmt.Println("[MASTER]\tAdding new client", sender_id)
                alive_timer := time.NewTimer(TIMEOUT_INTERVAL)
                client = com.Client {
                    ID: sender_id,
                    AliveTimer: alive_timer,
                }
                go ListenForClientTimeout(sender_id, alive_timer, client_timed_out)
            }
            client.AliveTimer.Reset(TIMEOUT_INTERVAL)
            client.HasTimedOut = false
            client.LastPassedFloor = data.LastPassedFloor
            clients[sender_id] = client

            orders = AddNewOrders(data.Requests, orders, sender_id)
            orders = DeleteDoneOrders(data.Requests, orders)


        case <- time_to_send.C:
            queue.DistributeWork(clients, orders)
            data := com.MasterData {
                AssignedBackup: backup,
                Orders:         orders,
                Clients:        clients,
            }
            c.ToClients <- network.Packet {
                Data: com.EncodeMasterData(data),
            }

        case who := <- client_timed_out:
            fmt.Println("[MASTER]\tClient", who, "timed out")
            client, exists := clients[who]
            if exists {
                RemoveExternalAssignments(orders, who)
                client.HasTimedOut = true
                clients[who] = client
            }
            if who == backup {
                WaitForBackup(c, orders, clients)
            }
        }
    }
}
