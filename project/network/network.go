package network

import (
    "net"
    "fmt"
    "log"
)

type ID string
const InvalidID ID = ""

type Packet struct {
    Address ID
    Data    []byte
}

// Might want to configure this at startup...
var client_port int = 10012
var master_port int = 20012

// TODO: Should we also include the port number?
// If so, return ID(sender.String()).
// It is needed to differentiate an ID from
// a machine that is both running a master instance
// and a client instance, but we currently don't
// need that.
func getSenderID(sender *net.UDPAddr) ID {
    return ID(sender.IP.String())
}

func GetMachineID() ID {
    ifaces, err := net.InterfaceAddrs()
    if err != nil {
        log.Fatal(err)
    }
    for _, addr := range(ifaces) {
        if ip_addr, ok := addr.(*net.IPNet); ok && !ip_addr.IP.IsLoopback() {
            if v4 := ip_addr.IP.To4(); v4 != nil {
                return ID(v4.String())
            }
        }
    }
    return "127.0.0.1"
}

func listen(socket *net.UDPConn, incoming chan Packet) {

    for {
        bytes := make([]byte, 1024)
        read_bytes, sender, err := socket.ReadFromUDP(bytes)
        if err == nil {
            incoming <- Packet{getSenderID(sender), bytes[:read_bytes]}
        } else {
            log.Println(err)
        }
    }
}

func broadcast(socket *net.UDPConn, to_port int, outgoing chan Packet) {

    bcast_addr := fmt.Sprintf("255.255.255.255:%d", to_port)
    remote, err := net.ResolveUDPAddr("udp", bcast_addr)
    if err != nil {
        log.Fatal(err)
    }
    for {
        packet := <- outgoing
        sent_bytes, err := socket.WriteToUDP(packet.Data, remote)
        if err != nil {
            log.Println(err)
        }
        fmt.Printf("[NETWORK]\tSent %d bytes to %d\n", sent_bytes, to_port)
    }
}

func bind(port int) *net.UDPConn {
    local, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port)) // Debugging

    // local, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatal(err)
    }

    socket, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }
    return socket
}

func ClientWorker(from_master, to_master chan Packet) {
    socket := bind(client_port)
    go listen(socket, from_master)
    broadcast(socket, master_port, to_master)
    socket.Close()
}

func MasterWorker(from_client, to_clients chan Packet) {
    socket := bind(master_port)
    go listen(socket, from_client)
    broadcast(socket, client_port, to_clients)
    socket.Close()
}
