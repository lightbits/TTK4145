package client

import (
    "../queue"
    "../driver"
    "../network"
    "../com"
    "../master"
    "time"
    "fmt"
    "log"
)

func WaitForMaster(c com.Channels, remaining_orders []com.Order) {
    fmt.Println("[CLIENT]\tWaiting for master")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case packet := <- c.FromMaster:
            ClientLoop(c, packet.Address)
            return

        case <- time_to_ping.C:
            c.ToMaster <- network.Packet {
                Data: []byte("Ping"),
            }

        case button := <- c.ButtonPressed:
            if button.Type == driver.ButtonOut {
                // TODO: Add to order list
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

func AcceptanceTestMasterData(data com.MasterData, requests []com.Order) bool {
    // Asserts that the master has gotten our latest requests
    for _, r := range(requests) {
        dirty := false
        for _, o := range(data.Orders) {
            if queue.IsSameOrder(r, o) && r.Done != o.Done {
                dirty = true
            }
        }
        if dirty {
            return false
        }
    }
    return true
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

func GetPriority(orders, requests []com.Order, our_id network.ID) int {
    for _, o := range(orders) {
        if o.TakenBy == our_id && o.Priority {
            return o.Button.Floor
        }
    }
    return driver.INVALID_FLOOR
}

// TODO: Pass LPF
func ClientLoop(c com.Channels, master_id network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    ORDER_DEADLINE_INTERVAL := 5 * driver.N_FLOORS * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    order_deadline := time.NewTimer(ORDER_DEADLINE_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    order_deadline.Stop()

    clients := make(map[network.ID]com.Client) // Local copy of master's client list
    orders := make([]com.Order, 0) // Local copy of master's queue
    requests := make([]com.Order, 0) // Unacknowledged local com

    our_id := network.GetMachineID()
    last_passed_floor := 0
    is_backup := false

    target_floor := driver.INVALID_FLOOR

    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
                go network.MasterWorker(c.FromClient, c.ToClients)
                go master.WaitForBackup(c, orders, clients)
            }

        case <- order_deadline.C:
            driver.MotorStop()
            log.Fatal("[FATAL]\tFailed to complete order within deadline.")

        case <- time_to_send.C:
            data := com.ClientData {
                LastPassedFloor: last_passed_floor,
                Requests:        requests,
            }
            c.ToMaster <- network.Packet {
                Data: com.EncodeClientData(data),
            }

        case button := <- c.ButtonPressed:
            requests = append(requests, com.Order {
                Button: button,
            })

        case floor := <- c.LastPassedFloorChanged:
            last_passed_floor = floor

        case floor := <- c.CompletedFloor:
            target_floor = driver.INVALID_FLOOR
            order_deadline.Stop()
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

            if data.AssignedBackup == our_id {
                is_backup = true
            } else {
                is_backup = false
            }

            orders = data.Orders
            SetButtonLamps(orders, our_id)

            new_target_floor := GetPriority(orders, requests, our_id)
            if target_floor != new_target_floor &&
                new_target_floor != driver.INVALID_FLOOR {

                target_floor = new_target_floor
                c.NewFloorOrder <- target_floor
                order_deadline.Reset(ORDER_DEADLINE_INTERVAL)
            }

            requests = RemoveAcknowledgedRequests(requests, orders)
        }
    }
}
