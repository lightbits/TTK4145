package network

import (
    "net"
    "fmt"
    "log"
)

type OutgoingPacket struct {
    Destination *net.UDPAddr
    Data []byte
}

type IncomingPacket struct {
    Sender   *net.UDPAddr
    Data     []byte
}

type ID *net.UDPAddr

func listen(socket *net.UDPConn, incoming chan IncomingPacket) {
    for {
        bytes := make([]byte, 1024)
        read_bytes, sender, err := socket.ReadFromUDP(bytes)
        if err != nil {
            log.Fatal(err)
        }
        incoming <- IncomingPacket{sender, bytes[:read_bytes]}
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

    broadcast, err := net.ResolveUDPAddr("udp", "127.0.0.255:20012")
    if err != nil {
        log.Fatal(err)
    }

    socket, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }
    defer socket.Close()

    go listen(socket, incoming)

    for {
        select {
        case p := <- outgoing:
            _, err = socket.WriteToUDP(p.Data, p.Destination)
            if err != nil {
                log.Fatal(err)
            }
        case p := <- outgoing_all:
            _, err = socket.WriteToUDP(p.Data, broadcast)
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
