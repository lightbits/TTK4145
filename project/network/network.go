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

const (
    tag_client_to_master byte = 0x00
    tag_master_to_all_clients byte = 0xff
)

func createTaggedPacket(data []byte, tag byte) []byte {
    protocol := []byte{byte(tag)}
    result := append(protocol[:], data[:]...)
    return result
}

func parseTaggedPacket(bytes []byte) ([]byte, byte) {
    tag := bytes[0]
    return bytes[1:], tag
}

func listen(
    socket *net.UDPConn,
    from_master chan IncomingPacket,
    from_client chan IncomingPacket) {

    for {
        bytes := make([]byte, 1024)
        read_bytes, sender, err := socket.ReadFromUDP(bytes)
        if err != nil {
            log.Println(err)
        } else {
            data, tag := parseTaggedPacket(bytes[:read_bytes])
            packet := IncomingPacket{ID(sender.String()), data}
            fmt.Printf("[NETWORK]\tRead %d bytes from %s tagged %x\n", read_bytes, sender.String(), tag)
            switch tag {
            case tag_client_to_master:
                from_client <- packet
            case tag_master_to_all_clients:
                from_master <- packet
            default:
                log.Println("Got an unknown tagged packet")
            }
        }
    }
}

func send(socket *net.UDPConn, destination string, data []byte, tag byte) {
    remote, err := net.ResolveUDPAddr("udp", destination)
    if err != nil {
        log.Fatal(err)
    }
    bytes := createTaggedPacket(data, tag)
    bytes_sent, err := socket.WriteToUDP(bytes, remote)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("[NETWORK]\tSent %d bytes to %s tagged %x\n", bytes_sent, destination, tag)
}

func GetMachineID(listen_port int) ID {
    return ID(fmt.Sprintf("127.0.0.1:%d", listen_port))
}

func Init(listen_port  int,
          to_master      chan OutgoingPacket,
          to_all_clients chan OutgoingPacket,
          to_any_master  chan OutgoingPacket,
          from_master    chan IncomingPacket,
          from_client    chan IncomingPacket) {
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

    go listen(socket, from_master, from_client)
    for {
        select {
        case packet := <- to_master:
            send(socket, string(packet.Destination), packet.Data, tag_client_to_master)

        case packet := <- to_all_clients:
            send(socket, broadcast.String(), packet.Data, tag_master_to_all_clients)

        case packet := <- to_any_master:
            send(socket, broadcast.String(), packet.Data, tag_client_to_master)
        }
    }
}
