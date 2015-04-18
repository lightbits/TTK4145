package master

import (
    "../queue"
    "../com"
    "../network"
    "../driver"
    "fmt"
    "time"
)

func WaitForBackup(c com.Channels, initial_queue []queue.Order,) {

    machine_id := network.GetMachineID()
    fmt.Println("[MASTER]\tRunning on machine", machine_id)
    fmt.Println("[MASTER]\tWaiting for backup...")
    for {
        select {
        case packet := <- c.FromClient:
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

func MasterLoop(c com.Channels, backup network.ID, initial_queue []queue.Order) {

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
            data := com.MasterData {
                AssignedBackup: backup,
                Orders:         orders,
            }
            c.ToClients <- network.Packet {
                Data: com.EncodeMasterData(data),
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