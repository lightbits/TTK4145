func TestNetwork(channels Channels) {
    go network.ClientWorker(channels.from_master, channels.to_master)
    go network.MasterWorker(channels.from_client, channels.to_clients)
    t1 := time.NewTimer(1*time.Second)
    t2 := time.NewTimer(2*time.Second)
    for {
        select {
        case <- t1.C:
            fmt.Println("[NETTEST]\tSending to all clients")
            channels.to_clients <- network.Packet{
                Data: []byte("A")}

        case <- t2.C:
            fmt.Println("[NETTEST]\tSending to any master")
            channels.to_master <- network.Packet{
                Data: []byte("BB")}

        case p := <- channels.from_client:
            fmt.Println("[NETTEST]\tClient sent:", len(p.Data), "bytes from", p.Address)

        case p := <- channels.from_master:
            fmt.Println("[NETTEST]\tMaster sent:", len(p.Data), "bytes from", p.Address)
        }
    }
}

func TestDriver(channels Channels) {
    for {
        select {
        case btn := <- channels.button_pressed:
            fmt.Println("[TEST]\tButton pressed")
            driver.SetButtonLamp(btn, true)
        case <- channels.floor_reached:
            fmt.Println("[TEST]\tFloor reached")
        case <- channels.stop_button:
            driver.ClearAllButtonLamps()
            fmt.Println("[TEST]\tStop button pressed")
        case <- channels.obstruction:
            fmt.Println("[TEST]\tObstruction changed")
        }
    }
}

func TestRemoveAcknowledgedRequests() bool {
    order := Order {
        Button: driver.OrderButton {
            Floor: 5,
            Type: driver.ButtonUp,
        },
    }
    orders := make([]Order, 0)
    orders = append(orders, order)
    order.Done = false
    order.Button.Floor = 5
    requests := make([]Order, 0)
    requests = append(requests, order)
    requests = RemoveAcknowledgedRequests(requests, orders)
    fmt.Println(requests)
    return len(requests) == 0
}
