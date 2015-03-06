package main

import (
    "fmt"
    "time"
    "./network"
)

/*
What does the client minimally need to send:
    * Keypresses
    * Timestamp (uint32 that is incremented per io event) per keypress
    * Last passed floor

What does the master minimally need to send to a client:
    * The next floor to go to
    * The button lamps to light
    * Acknowledged timestamps for io events????
      One for each?

Oh wait!!!!!!!

    Given the button lamps to light up, WE KNOW WHICH BUTTON PRESSES
    THE MASTER HAS REGISTERED!

    We don't need acknowledge numbers, the button lamps serve as
    ACKs!!!

    Then we don't need timestamps either!
*/

const CLIENT_UPDATE_INTERVAL = 1 * time.Second

func main() {
    outgoing := make(chan network.ClientUpdate)
    incoming := make(chan network.MasterUpdate)
    go network.InitClient(outgoing, incoming)

    ticker := time.NewTicker(CLIENT_UPDATE_INTERVAL)

    for {
        select {
        case <- ticker.C:
            fmt.Println("Client send update")
            outgoing <- network.ClientUpdate{Request: "Hello master!"}

        case update := <- incoming:
            fmt.Println("Master said:", update.ActiveOrders)
        }
    }
}
