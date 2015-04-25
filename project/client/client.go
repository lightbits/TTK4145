package client

import (
    "../queue"
    "../fakedriver"
    "../network"
    "../com"
    "../master"
    "../lift"
    "../logger"
    "time"
)

func IsSameOrderList(A, B []com.Order) bool {
    if len(A) != len(B) {
        return false
    }
    for _, a := range(A) {
        exists := false
        for _, b := range(B) {
            if queue.IsSameOrder(a, b) && a.Done == b.Done && a.TakenBy == b.TakenBy {
                exists = true
            }
        }
        if !exists {
            return false
        }
    }
    return true
}

func WaitForMaster(events           com.ClientEvents,
                   master_events    com.MasterEvents,
                   lift_events      com.LiftEvents) {

    println(logger.Info, "Waiting for master, Hello Sailor!")
    time_to_ping := time.NewTicker(1*time.Second)

    our_id := network.GetMachineID()
    orders := make([]com.Order, 0)

    for {
        select {
        case packet := <- events.FromMaster:
            println(logger.Debug, "Heard from a master")
            if len(orders) == 0 {
                remaining_orders := clientLoop(events, master_events, lift_events, packet.Address)
                println(logger.Info, "Waiting for master")
                for _, o := range(remaining_orders) {
                    if o.TakenBy == our_id {
                        orders = append(orders, o)
                    }
                }
                println(logger.Info, "Have remaining:", orders)
                queue.PrioritizeOrdersForSingleLift(orders, our_id, lift.GetLastPassedFloor())
                SetButtonLamps(orders, our_id)
                priority := queue.GetPriority(orders, our_id)
                if priority != nil {
                    lift_events.NewTargetFloor <- priority.Button.Floor
                }
            }

        case <- events.MissedDeadline:
            driver.MotorStop()
            println(logger.Fatal, "Failed to complete order within deadline")

        case floor := <- events.CompletedFloor:
            println(logger.Info, "Completed floor", floor)
            for i := 0; i < len(orders); i++ {
                if orders[i].TakenBy == our_id &&
                   orders[i].Button.Floor == floor  {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }
            queue.PrioritizeOrdersForSingleLift(orders, our_id, lift.GetLastPassedFloor())
            SetButtonLamps(orders, our_id)
            priority := queue.GetPriority(orders, our_id)
            if priority != nil {
                lift_events.NewTargetFloor <- priority.Button.Floor
            }

        case <- time_to_ping.C:
            println(logger.Debug, "Pinging")
            data := com.ClientData {
                LastPassedFloor: lift.GetLastPassedFloor(),
            }
            events.ToMaster <- network.Packet {
                Data: com.EncodeClientData(data),
            }

        case button := <- events.ButtonPressed:
            println(logger.Info, "Button pressed", button)
            if button.Type == driver.ButtonOut {
                orders = append(orders, com.Order {
                    Button:  button,
                    TakenBy: our_id,
                })
                queue.PrioritizeOrdersForSingleLift(orders, our_id, lift.GetLastPassedFloor())
                SetButtonLamps(orders, our_id)
                priority := queue.GetPriority(orders, our_id)
                if priority != nil {
                    lift_events.NewTargetFloor <- priority.Button.Floor
                }
            }
        }

    }
}

func RemoveAcknowledgedRequests(requests, orders []com.Order) []com.Order {
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

func SetButtonLamps(orders []com.Order, our_id network.ID) {
    driver.ClearAllButtonLamps()
    for _, o := range(orders) {
        if o.Button.Type == driver.ButtonOut && o.TakenBy != our_id {
            continue
        }
        driver.SetButtonLamp(o.Button, true)
    }
}

func clientLoop(events          com.ClientEvents,
                master_events   com.MasterEvents,
                lift_events     com.LiftEvents,
                master_id       network.ID) []com.Order {

    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    clients := make(map[network.ID]com.Client)
    orders := make([]com.Order, 0)
    requests := make([]com.Order, 0)

    our_id := network.GetMachineID()
    is_backup := false

    println(logger.Info, "Starting client with master", master_id)
    for {
        select {
        case <- master_timeout.C:
            println(logger.Info, "Master timed out")
            if is_backup {
                println(logger.Info, "We are backup, spawning new master!")
                go network.MasterWorker(master_events.FromClient, master_events.ToClients)
                go master.WaitForBackup(master_events, orders, clients)
            }
            return orders // Return remaining orders and wait for new master

        case <- events.MissedDeadline:
            driver.MotorStop()
            println(logger.Fatal, "Failed to complete order within deadline")

        case <- time_to_send.C:
            println(logger.Debug, "Sending")
            data := com.ClientData {
                LastPassedFloor: lift.GetLastPassedFloor(),
                Requests:        requests,
            }
            events.ToMaster <- network.Packet {
                Data: com.EncodeClientData(data),
            }

        case button := <- events.ButtonPressed:
            println(logger.Info, "Button pressed", button)
            requests = append(requests, com.Order {
                Button: button,
            })

        case floor := <- events.CompletedFloor:
            println(logger.Info, "Completed floor", floor)
            for _, o := range(orders) {
                if o.TakenBy == our_id && o.Button.Floor == floor {
                    o.Done = true
                    requests = append(requests, o)
                }
            }

        case packet := <- events.FromMaster:
            master_timeout.Reset(MASTER_TIMEOUT_INTERVAL)
            data, err := com.DecodeMasterPacket(packet.Data)
            if err != nil {
                break
            }
            println(logger.Debug, "Master said", data)
            clients = data.Clients
            if !IsSameOrderList(orders, data.Orders) {
                println(logger.Info, data.Orders)
            }
            orders = data.Orders
            is_backup = data.AssignedBackup == our_id
            SetButtonLamps(orders, our_id)

            priority := queue.GetPriority(orders, our_id)
            if priority != nil && !queue.IsOrderDone(*priority, requests) {
                lift_events.NewTargetFloor <- priority.Button.Floor
            }

            requests = RemoveAcknowledgedRequests(requests, orders)
        }
    }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "CLIENT", args...)
}
