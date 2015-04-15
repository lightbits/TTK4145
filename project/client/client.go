package client

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type channels struct {
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

            // Clear requests that are acknowledged
            for i := 0; i < len(requests); i++ {
                r := requests[i]
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
                    i--
                }
            }
        }
    }
}
