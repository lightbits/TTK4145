/*
Create a program (in any language, on any OS) that uses the process pair technique to print the numbers 1, 2, 3, 4, etc to a terminal window. The program should create its own backup: When the primary is running, only the primary should keep counting, and the backup should do nothing. When the primary dies, the backup should become the new primary, create its own new backup, and keep counting where the dead one left off. Make sure that no numbers are skipped!

You cannot rely on the primary telling the backup when it has died (because it would have to be dead first...). Instead, have the primary broadcast that it is alive a few times a second, and have the backup become the primary when a certain number of messages have been missed.
*/

/*
Solution:

The master process can be simplified to the following three actions

    a. Do internal work (state++)
    b. Do external work (print to display)
    c. Send state to backup

In which order should the master do these?

    abc
    acb
    bac
    bca
    cab
    cba

acb should work.

Assume the master state is at tick N.
Assume the backup has received state N.
Consider where the master might die:

    Before STATE++: Backup takes over from N.
                    PRINT N (duplicate)
                    STATE++
                    PRINT N+1 (new)

                    Ok. If we accept duplicates

    After STATE++
    Before SEND:    Backup takes over from N
                    PRINT N (duplicate)
                    STATE++
                    PRINT N+1

                    Ok.

    After SEND
    Before PRINT:   Backup takes over from N + 1
                    Print N+1 (new)

                    Ok.

    After PRINT     Backup takes over from N + 1
                    Print N+1 (duplicate)
                    STATE++
                    Print N+2 (new)

                    Ok



*/

package main

import (
    "time"
    "fmt"
    "strconv"
    "log"
    "net"
    "os"
    "os/exec"
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
    local, err := net.ResolveUDPAddr("udp", ":44556") // Change to 127.0.0.1 to work on laptop
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

func Log(str string) {
    fmt.Println(str)
    log.Println(str)
}

func LaunchBackupProcess() {
    // Windows
    // arg := "C:/Dokumenter/ttk4145/exercises/exercise6/start_backup.bat"
    // cmd := exec.Command("cmd", "/C", "start", arg)
    // err := cmd.Start()

    // Linux
    cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run backup.go")
    err := cmd.Run()

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

    // Set up com channel with backup
    outgoing_message := make(chan Message)
    go SendMessages(outgoing_message)

    LaunchBackupProcess()

    // Launch master process
    Log("Launching master process")
    state := NullState

    // Perhaps it is a restart?
    if len(os.Args) > 1 {
        initial_state, _ := strconv.Atoi(os.Args[1])
        state.Tick = int32(initial_state)
        Log(fmt.Sprintf("MASTER restart @%d", state.Tick))
        Log(fmt.Sprintf("MASTER PRINT %d", state.Tick))
    }

    for {
        Log("MASTER preparing work")
        time.Sleep(1 * time.Second)
        state.Tick++

        Log("MASTER finished work")
        time.Sleep(1 * time.Second)
        outgoing_message <- Message{state}

        Log("MASTER sent state to backup")
        time.Sleep(1 * time.Second)

        Log(fmt.Sprintf("MASTER PRINT %d", state.Tick))
        Log("")
        time.Sleep(1 * time.Second)
    }
}
