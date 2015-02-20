/*
Create a program (in any language, on any OS) that uses the process pair technique to print the numbers 1, 2, 3, 4, etc to a terminal window. The program should create its own backup: When the primary is running, only the primary should keep counting, and the backup should do nothing. When the primary dies, the backup should become the new primary, create its own new backup, and keep counting where the dead one left off. Make sure that no numbers are skipped!

You cannot rely on the primary telling the backup when it has died (because it would have to be dead first...). Instead, have the primary broadcast that it is alive a few times a second, and have the backup become the primary when a certain number of messages have been missed.
*/

package main

import (
    "time"
    "fmt"
    "strconv"
    "log"
    "net"
    "os"
    "encoding/binary"
    "bytes"
)

type State struct {
    Tick int32
}

type Message struct {
    PrimaryState State
}

var NullState State = State{0}

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
    }
}

func main() {
    fmt.Println("Launching master process")

    outgoing_message := make(chan Message)
    go SendMessages(outgoing_message)

    // TODO: Launch backup automatically

    state := NullState

    if len(os.Args) > 1 {
        initial_state, _ := strconv.Atoi(os.Args[1])
        state.Tick = int32(initial_state)
    }

    for {
        fmt.Println("MASTER preparing work")
        time.Sleep(1 * time.Second)
        state.Tick++

        fmt.Println("MASTER finished work")
        time.Sleep(1 * time.Second)
        outgoing_message <- Message{state}

        fmt.Println("MASTER sent state to backup")
        time.Sleep(1 * time.Second)

        fmt.Println("MASTER PRINT", state.Tick)
        fmt.Println()
        time.Sleep(1 * time.Second)
    }
}
