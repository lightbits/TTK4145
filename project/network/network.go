package network

import (
    "net"
    "fmt"
    "log"
)

type ID string
const InvalidID ID = ""

type packet struct {
    Address ID
    Data    []byte
}

type ClientData struct {
    LastPassedFloor int
    Requests        []Order
}

type MasterData struct {
    AssignedBackup network.ID
    Orders         []Order
}

type ClientEvents struct {
    to_master       chan network.Packet
    from_master     chan network.Packet
}

type MasterEvents struct {
    to_clients       chan network.Packet
    from_client      chan network.Packet
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

func DecodeMasterPacket(b []byte) (MasterData, error) {
    var result MasterData
    err := json.Unmarshal(b, &result)
    if err == nil {
        for _, o := range(result.Orders) {
            if o.TakenBy == network.InvalidID {
                log.Fatal("[CLIENT]\tA non-taken order was received")
            }
        }
    }

    return result, err
}

func DecodeClientPacket(b []byte) (ClientData, error) {
    var result ClientData
    err := json.Unmarshal(b, &result)
    return result, err
}

func EncodeMasterData(m MasterData) []byte {
    result, err := json.Marshal(m)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func EncodeClientData(c ClientData) []byte {
    result, err := json.Marshal(c)
    if err != nil {
        log.Fatal(err)
    }
    return result
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
    // local, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port)) // Debugging

    local, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatal(err)
    }

    socket, err := net.ListenUDP("udp", local)
    if err != nil {
        log.Fatal(err)
    }
    return socket
}

func ClientWorker(from_master chan MasterData, to_master chan ClientData) {
    socket := bind(client_port)
    go listen(socket, from_master)
    broadcast(socket, master_port, to_master)
    socket.Close()
}

func MasterWorker(from_client chan ClientData, to_clients chan MasterData) {
    socket := bind(master_port)
    go listen(socket, from_client)
    broadcast(socket, client_port, to_clients)
    socket.Close()
}
