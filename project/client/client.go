package client

import (
    "../queue"
    "../fakedriver"
    "../network"
    "../com"
    "../master"
    "../lift"
    "time"
    "fmt"
    "log"
)

func WaitForMaster(c com.Channels, remaining_orders []com.Order) {
    fmt.Println("[CLIENT]\tWaiting for master")
    time_to_ping := time.NewTicker(1*time.Second)

    orders := remaining_orders
    our_id := network.GetMachineID()

    for {
        select {
        case packet := <- c.FromMaster:
            if len(orders) == 0 {
                ClientLoop(c, packet.Address)
                return
            }

        case <- c.MissedDeadline:
            driver.MotorStop()
            log.Fatal("[FATAL]\tFailed to complete order within deadline.")

        case floor := <- c.CompletedFloor:
            for i := 0; i < len(orders); i++ {
                if orders[i].TakenBy == our_id &&
                   orders[i].Button.Floor == floor  {
                    orders = append(orders[:i], orders[i+1:]...)
                }
            }
            queue.PrioritizeOrdersForSingleLift(orders, our_id, lift.GetLastPassedFloor())
            SetButtonLamps(orders, our_id)
            c.NewOrders <- orders

        case <- time_to_ping.C:
            c.ToMaster <- network.Packet {
                Data: []byte("Ping"),
            }

        case button := <- c.ButtonPressed:
            if button.Type == driver.ButtonOut {
                orders = append(orders, com.Order {
                    Button:  button,
                    TakenBy: our_id,
                })
                queue.PrioritizeOrdersForSingleLift(orders, our_id, lift.GetLastPassedFloor())
                SetButtonLamps(orders, our_id)
                c.NewOrders <- orders
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

func ClientLoop(c com.Channels, master_id network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    clients := make(map[network.ID]com.Client) // Local copy of master's client list
    orders := make([]com.Order, 0) // Local copy of master's queue
    requests := make([]com.Order, 0) // Unacknowledged local com

    our_id := network.GetMachineID()
    is_backup := false

    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
                go network.MasterWorker(c.FromClient, c.ToClients)
                go master.WaitForBackup(c, orders, clients)
            } else {
                WaitForMaster(c, orders)
                return
            }

        case <- c.MissedDeadline:
            driver.MotorStop()
            log.Fatal("[FATAL]\tFailed to complete order within deadline.")

        case <- time_to_send.C:
            data := com.ClientData {
                LastPassedFloor: lift.GetLastPassedFloor(),
                Requests:        requests,
            }
            c.ToMaster <- network.Packet {
                Data: com.EncodeClientData(data),
            }

        case button := <- c.ButtonPressed:
            requests = append(requests, com.Order {
                Button: button,
            })

        case floor := <- c.CompletedFloor:
            for _, o := range(orders) {
                if o.TakenBy == our_id && o.Button.Floor == floor {
                    o.Done = true
                    requests = append(requests, o)
                }
            }

        case packet := <- c.FromMaster:
            master_timeout.Reset(MASTER_TIMEOUT_INTERVAL)
            data, err := com.DecodeMasterPacket(packet.Data)
            if err != nil {
                break
            }
            fmt.Println("[CLIENT]\tMaster said", data)
            clients = data.Clients
            orders = data.Orders
            is_backup = data.AssignedBackup == our_id
            SetButtonLamps(orders, our_id)
            c.NewOrders <- orders
            requests = RemoveAcknowledgedRequests(requests, orders)
        }
    }
}
