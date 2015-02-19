/*
Create a program (in any language, on any OS) that uses the process pair technique to print the numbers 1, 2, 3, 4, etc to a terminal window. The program should create its own backup: When the primary is running, only the primary should keep counting, and the backup should do nothing. When the primary dies, the backup should become the new primary, create its own new backup, and keep counting where the dead one left off. Make sure that no numbers are skipped!

You cannot rely on the primary telling the backup when it has died (because it would have to be dead first...). Instead, have the primary broadcast that it is alive a few times a second, and have the backup become the primary when a certain number of messages have been missed.
*/

package main

import (
    // "os/exec"
    "time"
    "log"
    "net"
    "encoding/binary"
    "bytes"
    "fmt"
)

type State struct {
    Tick uint32
}

type Message struct {
    PrimaryState State
}

func ListenForMessages(incoming_message chan Message) {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:33445")
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
        // log.Println("Read", read_bytes, "from", sender)

        b := bytes.NewBuffer(buffer)
        m := Message{}
        binary.Read(b, binary.BigEndian, &m)
        incoming_message <- m
    }
}

func SendMessages(outgoing_message chan Message) {
    local, err := net.ResolveUDPAddr("udp", "127.0.0.1:44556")
    if err != nil {
        log.Fatal(err)
    }

    bcast, err := net.ResolveUDPAddr("udp", "255.255.255.255:33445")
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    for {
        m := <- outgoing_message
        buffer := &bytes.Buffer{}
        binary.Write(buffer, binary.BigEndian, m)

        _, err := conn.WriteToUDP(buffer.Bytes(), bcast)
        if err != nil {
            log.Fatal(err)
        }
        // log.Println("Sent", sent_bytes)
        time.Sleep(1 * time.Second)
    }
}

func Master(initial_state State) {
    outgoing_message := make(chan Message)
    go SendMessages(outgoing_message)

    // Startup
    time.Sleep(2 * time.Second)

    state := initial_state
    for {
        // Pretend to do work
        state.Tick++
        fmt.Println("PRINT", state.Tick)
        // ...

        // Send state update to backup every second
        outgoing_message <- Message{state}
        time.Sleep(1 * time.Second)

        if (state.Tick == 10) {
            time.Sleep(5 * time.Second)
        }
    }
}

func Backup(null_state State) {
    incoming_message := make(chan Message)
    go ListenForMessages(incoming_message)

    state := null_state
    for {
        select {
        case <-time.After(3 * time.Second):
            // Take over
            log.Println("Primary loss detected. Take over @", state.Tick)
        case msg := <- incoming_message:
            state = msg.PrimaryState
            log.Println("Update received. Primary state @", state.Tick)
        }
    }
}

func main() {
    null_state := State{0}
    go Master(null_state)
    go Backup(null_state)

    time.Sleep(20 * time.Second)
}
