/*
Create a program (in any language, on any OS) that uses the process pair technique to print the numbers 1, 2, 3, 4, etc to a terminal window. The program should create its own backup: When the primary is running, only the primary should keep counting, and the backup should do nothing. When the primary dies, the backup should become the new primary, create its own new backup, and keep counting where the dead one left off. Make sure that no numbers are skipped!

You cannot rely on the primary telling the backup when it has died (because it would have to be dead first...). Instead, have the primary broadcast that it is alive a few times a second, and have the backup become the primary when a certain number of messages have been missed.
*/

package main

import (
    // "os/exec"
    "time"
    "net"
    "log"
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

func SendMessages(outgoing_message chan Message) {
    local, err := net.ResolveUDPAddr("udp", ":44556")
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
        // fmt.Println("Sent", sent_bytes)
        time.Sleep(1 * time.Second)
    }
}

func Master(initial_state State, 
            outgoing_message chan Message,
            incoming_message chan Message) {

    fmt.Println("Launching master process")
    // TODO: Launch backup automatically

    state := initial_state
    for {
        // Pretend to do work
        fmt.Println("MASTER preparing work")
        time.Sleep(1 * time.Second)
        state.Tick++
        fmt.Println("MASTER finished work")
        time.Sleep(1 * time.Second)
        outgoing_message <- Message{state}
        fmt.Println("MASTER sent state to backup")
        time.Sleep(1 * time.Second)
        // ...

        // Send state update to backup every second
        fmt.Println("MASTER PRINT", state.Tick)
        fmt.Println()
        time.Sleep(1 * time.Second)

        if (state.Tick == 5) {
            time.Sleep(8 * time.Second)
            break
        }
    }
}

func Backup(null_state State, outgoing_message chan Message, incoming_message chan Message) {
    fmt.Println("Launching backup process")
    state := null_state

    Loop:
    for {
        select {
        case <-time.After(7 * time.Second):
            fmt.Println("BACKUP Primary loss detected. Take over @", state.Tick)

            // We don't know if the master was able to do this yet, so we do it
            // in the risk of duplicate prints
            fmt.Println("BACKUP PRINT", state.Tick)

            go Master(state, outgoing_message, incoming_message)
            break Loop
        case msg := <- incoming_message:
            state = msg.PrimaryState
            fmt.Println("BACKUP Update received. Primary state @", state.Tick)
        }
    }
}

func main() {
    incoming_message := make(chan Message)
    outgoing_message := make(chan Message)
    go ListenForMessages(incoming_message)
    go SendMessages(outgoing_message)
    null_state := State{0}

    go Master(null_state, outgoing_message, incoming_message)
    // go Backup(null_state)

    time.Sleep(40 * time.Second)
}
