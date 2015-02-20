package main

import (
    "time"
    "net"
    "log"
    "os"
    "os/exec"
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
    local, err := net.ResolveUDPAddr("udp", ":33445") // Change to 127.0.0.1 to work on laptop
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
        b := bytes.NewBuffer(buffer)
        m := Message{}
        binary.Read(b, binary.BigEndian, &m)
        incoming_message <- m
    }
}

func RestartMasterProcess(initial_state State) {
    // Linux
    arg := fmt.Sprintf("go run primary.go %d", initial_state.Tick)
    cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", arg)
    err := cmd.Run()

    // Windows
    // arg0 := "C:/Dokumenter/ttk4145/exercises/exercise6/start_primary.bat"
    // arg1 := fmt.Sprintf("%d", initial_state.Tick)
    // cmd := exec.Command("cmd", "/C", "start", arg0, arg1)
    // err := cmd.Start()

    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // Redirect log output
    f, err := os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    defer f.Close()
    log.SetOutput(f)

    // Create com channel with master
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
            RestartMasterProcess(state)
            return
        case msg := <- incoming_message:
            state = msg.PrimaryState
            fmt.Println("BACKUP Update received. Primary state @", state.Tick)
        }
    }
}
