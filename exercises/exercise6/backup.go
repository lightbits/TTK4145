package main

import (
    "time"
    "net"
    "log"
    "encoding/binary"
    "bytes"
    "fmt"
)

type State struct {
    Tick int32
}

type Message struct {
    PrimaryState State
}

var NullState State = State{0}

func ListenForMessages(incoming_message chan Message) {
    local, err := net.ResolveUDPAddr("udp", ":33445")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    for {
        buffer := make([]byte, 1024)
        _, _, err := conn.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal(err)
        }
        // fmt.Println("Read", read_bytes, "from", sender)

        b := bytes.NewBuffer(buffer)
        m := Message{}
        binary.Read(b, binary.BigEndian, &m)
        incoming_message <- m
    }
}

func main() {
    incoming_message := make(chan Message)
    go ListenForMessages(incoming_message)
    
    fmt.Println("Launching backup process")
    state := NullState

    // Wait for initial state
    select {
    case msg := <- incoming_message:
        state = msg.PrimaryState
        fmt.Println("BACKUP Received initial state update @", state.Tick)
    case <- time.After(7 * time.Second):
        fmt.Println("BACKUP Has primary not started?")
    }

    for {
        select {
        case <- time.After(7 * time.Second):
            fmt.Println("BACKUP Primary loss detected. Take over @", state.Tick)
            fmt.Println("BACKUP PRINT", state.Tick)
            return
        case msg := <- incoming_message:
            state = msg.PrimaryState
            fmt.Println("BACKUP Update received. Primary state @", state.Tick)
        }
    }
}
