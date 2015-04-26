package master

import (
    "../queue"
    "../com"
    "../network"
    "../driver"
    "../logger"
    "time"
)

const client_timeout_period = 5 * time.Second
const net_send_period = 250 * time.Millisecond
const can_use_self_as_backup_period = 10 * time.Second

func WaitForBackup(events com.MasterEvents,
                   initial_queue []com.Order,
                   initial_clients map[network.ID]com.Client) {

    machine_id := network.GetMachineID()
    println(logger.Info, "Waiting for backup on machine", machine_id)

    can_use_self_as_backup_timer := time.NewTimer(can_use_self_as_backup_period)
    can_use_self_as_backup := false

    queue := initial_queue
    clients := initial_clients

    for {
        select {
        case <- can_use_self_as_backup_timer.C:
            can_use_self_as_backup = true

        case packet := <- events.FromClient:
            _, err := com.DecodeClientPacket(packet.Data)
            if err != nil {
                break
            }

            if (packet.Address == machine_id && can_use_self_as_backup) ||
               (packet.Address != machine_id) {
                if packet.Address == machine_id {
                    println(logger.Info, "Using self as backup!")
                }
                queue, clients = masterLoop(events, packet.Address, queue, clients)
                println(logger.Info, "Waiting for backup on machine", machine_id)
                println(logger.Info, "Have queue:", queue)
                println(logger.Info, "Have clients:", clients)
                can_use_self_as_backup_timer = time.NewTimer(can_use_self_as_backup_period)
            }
        }
    }
}

func listenForClientTimeout(id network.ID, timer *time.Timer, timeout chan network.ID) {
    for {
        select {
        case <- timer.C:
            timeout <- id
        }
    }
}

func addNewOrders(requests, orders []com.Order, sender network.ID) []com.Order {
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

func deleteDoneOrders(requests, orders []com.Order) []com.Order {
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

func masterLoop(events com.MasterEvents,
                backup network.ID,
                initial_queue []com.Order,
                initial_clients map[network.ID]com.Client) ([]com.Order,
                map[network.ID]com.Client) {

    time_to_send := time.NewTicker(net_send_period)
    client_timed_out := make(chan network.ID)

    orders := initial_queue
    if initial_queue == nil {
        orders = make([]com.Order, 0)
    }

    clients := make(map[network.ID]com.Client)
    if initial_clients != nil {
        for _, c := range(initial_clients) {
            c.AliveTimer = time.NewTimer(client_timeout_period)
            clients[c.ID] = c
            go listenForClientTimeout(c.ID, c.AliveTimer, client_timed_out)
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
                alive_timer := time.NewTimer(client_timeout_period)
                client = com.Client {
                    ID: sender_id,
                    AliveTimer: alive_timer,
                }
                go listenForClientTimeout(sender_id, alive_timer, client_timed_out)
            }
            println(logger.Debug, "Resetting", packet.Address, "'s timer")
            client.AliveTimer.Reset(client_timeout_period)
            client.HasTimedOut = false
            client.LastPassedFloor = data.LastPassedFloor
            clients[sender_id] = client

            orders = addNewOrders(data.Requests, orders, sender_id)
            orders = deleteDoneOrders(data.Requests, orders)


        case <- time_to_send.C:
            println(logger.Debug, "Sending to clients")
            err := queue.DistributeWork(clients, orders)
            if err != nil {
                println(logger.Fatal, err)
            }
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
                client.HasTimedOut = true
                clients[who] = client
                err := queue.DistributeWork(clients, orders)
                if err != nil {
                    println(logger.Fatal, err)
                }
            }
            if who == backup {
                return orders, clients // Return state and wait for new backup
            }
        }
    }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "MASTER", args...)
}
