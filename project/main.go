package main

import (
    "log"
    "net"
    "time"
)

type Message struct {
    Sender *net.UDPAddr
    Data string
}

type Connection struct {
    Address *net.UDPAddr
}

func Listen(done chan bool, 
    connection_channel chan Connection,
    message_channel chan Message) {
    // The address we wish to listen to
    local, err := net.ResolveUDPAddr("udp", ":20012")
    if err != nil {
        log.Fatal(err)
    }

    // Create a socket
    conn, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }

    // Read forever
    for {
        buffer := make([]byte, 1024)
        _, sender, err := conn.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal(err)
        }

        data := string(buffer)
        if (data[0] == 'p') {
            connection_channel <- Connection{sender}
            time.Sleep(1 * time.Second)
            conn.WriteToUDP([]byte("Hello to you!"), sender)
        } else if (data[0] == 'm') {
            message_channel <- Message{sender, data}   
        }
        time.Sleep(1 * time.Second)
    }

    done <- true
}

// func Write(done chan bool) {
//     // Server address
//     remote, err := net.ResolveUDPAddr("udp", "129.241.187.255:20012")
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Create a "connection" socket, use arbitrary local port
//     conn, err := net.DialUDP("udp", nil, remote)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Send forever
//     for {
//         // Try to send something
//         buffer := []byte("Hoopdoopawdop")
//         bytes_sent, err := conn.Write(buffer)
//         if err != nil {
//             log.Fatal(err)
//         }
//         log.Println("Sent", bytes_sent)
//         time.Sleep(1 * time.Second)
//     }

//     done <- true
// }

func main() {
    done := make(chan bool)
    message_channel := make(chan Message)
    connection_channel := make(chan Connection)

    go Listen(done, connection_channel, message_channel)

    for {
        select {
        case message := <-message_channel:
            log.Println(message.Sender, "said", message.Data)
        case connection := <-connection_channel:
            // TODO: Check if new connection or existing connection is pinging us
            log.Println(connection, "wants to connect or is pinging us!")
        }
    }
    
    <-done
}

/*
network module
---------------

struct Connection {
    NetAddress (ip and port)
    LastPingTime
}

Connection active_connections

func ListenForConnections() {
    
    connection, sender = listen_socket.accept()
    active_connections[sender].LastPingTime = time.now()
}

*/

// func NetSendToAll()

// // #cgo LDFLAGS: -L./driver -lelev -lcomedi -lm
// // #include "driver/elev.h"
// import "C"

// func ElevInit() {
//     C.elev_init()
// }

// func main() {
//     C.elev_init()
//     C.elev_set_motor_direction(-1)
//     time.Sleep(1 * time.Second)
//     C.elev_set_motor_direction(0)

//     // for {
//     //     var floor C.int = C.elev_get_floor_sensor_signal()
//     //     if floor == 3 {
//     //         C.elev_set_motor_direction(-1)
//     //     } else if floor == 0 {
//     //         C.elev_set_motor_direction(1)
//     //     }
//     // }
//     fmt.Println("Hey!")
// }