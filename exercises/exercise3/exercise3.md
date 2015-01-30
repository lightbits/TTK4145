UDP is actually connectionless, i.e. we only need our local
address to create a UDP socket. When sending a data over the
socket we would then specify the destination address each time.

Go provides us with a convenience feature around this, namely

    local, _  := net.ResolveUDPAddr("udp", "<our-ip>:<our listen port>")
    remote, _ := net.ResolveUDPAddr("udp", "<server-ip>:<server listen port>")
    conn, _   := net.DialUDP("udp", local, remote)

When creating a socket like this, the remote address is stored
together with the returned conn. When we wish to send data we
can then simply write

    data := []byte("Testing 123")
    conn.Write(data)

If we instead want to go the traditional route, we must create
the socket like so

    conn, _ := net.ListenUDP("udp", local)

and to send data we must do

    conn.WriteToUDP(data, remote)

Naturally, it is an error to use WriteToUDP with a socket created
with DialUDP.

Resources
---------
https://golang.org/src/net/udp_test.go