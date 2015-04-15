package master

type client struct {
    ID              network.ID
    LastPassedFloor int
    Timer           *time.Timer
    HasTimedOut     bool
    // TODO: Add timer which fires off if the client has been on the same
    // floor for a long time, when it has a floor to go to.
}

type channels struct {
    to_clients      chan network.Packet
    from_client     chan network.Packet
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
            // MasterLoop(c, packet.Address)

            if packet.Address != machine_id {
                MasterLoop(c, packet.Address, initial_queue)
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
