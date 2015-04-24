package master

import (
    "../queue"
    "../com"
    "../network"
    "../fakedriver"
    "../logger"
    "time"
)

func WaitForBackup(events          com.MasterEvents,
                   initial_queue   []com.Order,
                   initial_clients map[network.ID]com.Client) {

    machine_id := network.GetMachineID()
    go network.MasterWorker(events.FromClient, events.ToClients)
    println(logger.Info, "Waiting for backup on machine", machine_id)
    println(logger.Info, "Initial queue:", initial_queue)
    println(logger.Info, "Initial clients:", initial_clients)
    for {
        select {
        case packet := <- events.FromClient:
            _, err := com.DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }
            // DEBUG:
            // MasterLoop(events, packet.Address, initial_queue, initial_clients)
            // return

            if packet.Address != machine_id {
                MasterLoop(events, packet.Address, initial_queue, initial_clients)
                return
            } else {
                println(logger.Debug, "Cannot use own machine as backup client")
            }
        }
    }
}

func ListenForClientTimeout(id network.ID, timer *time.Timer, timeout chan network.ID) {
    for {
        select {
        case <- timer.C:
            timeout <- id
        }
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

func MasterLoop(events          com.MasterEvents,
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

    println(logger.Info, "Starting with backup", backup)
    for {
        select {
        case packet := <- events.FromClient:
            data, err := com.DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }
            println(logger.Debug, "Client said", data)

            sender_id := packet.Address
            client, exists := clients[sender_id]
            if !exists {
                println(logger.Info, "Adding new client", sender_id)
                alive_timer := time.NewTimer(TIMEOUT_INTERVAL)
                client = com.Client {
                    ID: sender_id,
                    AliveTimer: alive_timer,
                }
                go ListenForClientTimeout(sender_id, alive_timer, client_timed_out)
            }
            println(logger.Debug, "Resetting", packet.Address, "'s timer")
            client.AliveTimer.Reset(TIMEOUT_INTERVAL)
            client.HasTimedOut = false
            client.LastPassedFloor = data.LastPassedFloor
            clients[sender_id] = client

            orders = AddNewOrders(data.Requests, orders, sender_id)
            orders = DeleteDoneOrders(data.Requests, orders)


        case <- time_to_send.C:
            println(logger.Debug, "Sending to clients")
            queue.DistributeWork(clients, orders)
            data := com.MasterData {
                AssignedBackup: backup,
                Orders:         orders,
                Clients:        clients,
            }
            events.ToClients <- network.Packet {
                Data: com.EncodeMasterData(data),
            }

        case who := <- client_timed_out:
            println(logger.Info, who, "timed out")
            client, exists := clients[who]
            if exists {
                RemoveExternalAssignments(orders, who)
                client.HasTimedOut = true
                clients[who] = client
            }
            if who == backup {
                WaitForBackup(events, orders, clients)
            }
        }
    }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "MASTER", args...)
}