package network

import (
    "net"
    "fmt"
    "log"
)

type ID string

type OutgoingPacket struct {
    Destination ID
    Data        []byte
}

type IncomingPacket struct {
    Sender   ID
    Data     []byte
}

func listen(socket *net.UDPConn, incoming chan IncomingPacket) {
    for {
        bytes := make([]byte, 1024)
        read_bytes, sender, err := socket.ReadFromUDP(bytes)
        if err != nil {
            log.Println(err)
        } else {
            incoming <- IncomingPacket{ID(sender.String()), bytes[:read_bytes]}
        }
    }
}

func Init(listen_port  int,
          outgoing     chan OutgoingPacket,
          outgoing_all chan OutgoingPacket,
          incoming     chan IncomingPacket) {
    local, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", listen_port))
    if err != nil {
        log.Fatal(err)
    }

    // TODO: Test broadcasting at lab
    broadcast, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", listen_port))
    if err != nil {
        log.Fatal(err)
    }

    // socket_all, err := net.DialUDP("udp", local, broadcast)
    // if err != nil {
    //     log.Fatal(err)
    // }

    socket, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }
    defer socket.Close()

    go listen(socket, incoming)
    for {
        select {
        case p := <- outgoing:
            remote, err := net.ResolveUDPAddr("udp", string(p.Destination))
            if err != nil {
                log.Fatal(err)
            }
            bytes_sent, err := socket.WriteToUDP(p.Data, remote)
            if err != nil {
                log.Fatal(err)
            }
            log.Println("Sent", bytes_sent, "bytes to", p.Destination)

        case p := <- outgoing_all:
            bytes_sent, err := socket.WriteToUDP(p.Data, broadcast)
            // bytes_sent, err := socket_all.Write(p.Data)
            if err != nil {
                log.Fatal(err)
            }
            log.Println("Broadcasted", bytes_sent, "bytes")
        }
    }
}
