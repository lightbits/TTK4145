package client

import (
    "../queue"
    "../driver"
    "../network"
    "../com"
    "../master"
    "time"
    "fmt"
)

func WaitForMaster(c com.Channels, remaining_orders []com.Order) {
    fmt.Println("[CLIENT]\tWaiting for ..")
    time_to_ping := time.NewTicker(1*time.Second)

    for {
        select {
        case packet := <- c.FromMaster:
            fmt.Println("[CLIENT]\tHeard a master!")
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
        case <- c.FloorReached:
        case <- c.StopButton: // ignore
        case <- c.Obstruction: // ignore
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

func ClientLoop(c com.Channels, master_id network.ID) {
    MASTER_TIMEOUT_INTERVAL := 5 * time.Second
    SEND_INTERVAL := 250 * time.Millisecond

    master_timeout := time.NewTimer(MASTER_TIMEOUT_INTERVAL)
    time_to_send := time.NewTicker(SEND_INTERVAL)

    clients := make(map[network.ID]com.Client) // Local copy of master's client list
    orders := make([]com.Order, 0) // Local copy of master's queue
    requests := make([]com.Order, 0) // Unacknowledged local com

    our_id := network.GetMachineID()
    last_passed_floor := 0
    is_backup := false

    fmt.Println("[CLIENT]\tStarting client")
    for {
        select {
        case <- master_timeout.C:
            if is_backup {
                fmt.Println("[CLIENT]\tMaster timed out; taking over!")
                fmt.Println("[CLIENT]\tUsing:", clients)

                go network.MasterWorker(c.FromClient, c.ToClients)
                go master.WaitForBackup(c, orders, clients)
            }

        case <- time_to_send.C:
            data := com.ClientData {
                LastPassedFloor: last_passed_floor,
                Requests:        requests,
            }
            c.ToMaster <- network.Packet {
                Data: com.EncodeClientData(data),
            }

        case button := <- c.ButtonPressed:
            fmt.Println("[CLIENT]\tA button was pressed")
            order := com.Order {
                Button: button,
            }
            requests = append(requests, order)

        case floor := <- c.LastPassedFloorChanged:
            last_passed_floor = floor

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

            if data.AssignedBackup == our_id {
                is_backup = true
            } else {
                is_backup = false
            }

            driver.ClearAllButtonLamps()

            // TODO: Make this cleaner/more readable/understandable
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
                        c.TargetFloorChanged <- o.Button.Floor
                    }
                    fmt.Println("[CLIENT]\tTarget floor:", o.Button.Floor)
                }
            }

            requests = RemoveAcknowledgedRequests(requests, orders)
        }
    }
}
