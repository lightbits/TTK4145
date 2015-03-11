package main

import (
    "fmt"
    "time"
    "net"
    "log"
    "encoding/json"
)

const CLIENT_UPDATE_INTERVAL = 1 * time.Second

type Status struct {
    LastPassedFloor int
    ClearedFloors   []int
    Commands        []OrderButton
}

type ButtonType int
const (
    ButtonUp ButtonType = iota
    ButtonDown
    ButtonOut
)

type OrderButton struct {
    Floor int
    Type  ButtonType
}

type MasterUpdate struct {
    LitButtonLamps []OrderButton
    TargetFloor    int
}

func sendToMaster(conn *net.UDPConn, status Status) {
    remote, err := net.ResolveUDPAddr("udp", "127.0.0.1:20012")
    if err != nil {
        log.Fatal(err)
    }
    bytes, err := json.Marshal(status)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.WriteToUDP(bytes, remote)
    if err != nil {
        log.Fatal(err)
    }
}

func listenToMaster(conn *net.UDPConn, incoming chan MasterUpdate) {
    for {
        bytes := make([]byte, 1024)
        read_bytes, _, err := conn.ReadFromUDP(bytes)
        if err != nil {
            log.Fatal(err)
        }

        var update MasterUpdate
        err = json.Unmarshal(bytes[:read_bytes], &update)
        if err != nil{
            log.Fatal(err)
        }
        incoming <- update
    }
}

func fakeEventGenerator(
    reached_floor chan int,
    cleared_floor chan int,) {
    time.Sleep(5 * time.Second)
        reached_floor <- 5
        time.Sleep(3 * time.Second)
        cleared_floor <- 5
}

func main() {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:54321")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    incoming_update := make(chan MasterUpdate)
    time_to_send    := time.NewTicker(CLIENT_UPDATE_INTERVAL)
    reached_floor   := make(chan int)
    cleared_floor   := make(chan int)
    go listenToMaster(conn, incoming_update)
    go fakeEventGenerator(reached_floor, cleared_floor)

    var status Status

    for {
        select {
        case f := <- reached_floor:
            status.LastPassedFloor = f

        case f := <- cleared_floor:
            status.ClearedFloors = append(status.ClearedFloors, f)

        case <- time_to_send.C:
            fmt.Println("Client send update")
            status.Commands = []OrderButton{
                OrderButton{5, ButtonUp},
            }
            sendToMaster(conn, status)

        case update := <- incoming_update:
            fmt.Println("Master said:", update.LitButtonLamps, update.TargetFloor)
        }
    }
}
