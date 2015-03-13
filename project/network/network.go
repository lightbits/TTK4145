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
            log.Println(err)
        } else {
            incoming <- IncomingPacket{sender, bytes[:read_bytes]}
        }
    }
}

func Init(listen_port  int,
          outgoing     chan OutgoingPacket,
          outgoing_all chan OutgoingPacket,
          incoming     chan IncomingPacket) {
    local, err := net.ResolveUDPAddr("udp", fmt.Sprintf("78.91.19.229:%d", listen_port))
    if err != nil {
        log.Fatal(err)
    }

    broadcast, err := net.ResolveUDPAddr("udp", "255.255.255.255:20012")
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
            bytes_sent, err = socket.WriteToUDP(p.Data, p.Destination)
            if err != nil {
                log.Fatal(err)
            }
            log.Println("Sent", bytes_sent, "bytes to", p.Destination)

        case p := <- outgoing_all:
            bytes_sent, err = socket.WriteToUDP(p.Data, broadcast)
            if err != nil {
                log.Fatal(err)
            }
            log.Println("Broadcasted", bytes_sent, "bytes")
        }
    }
}
