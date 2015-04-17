package queue

import (
	"log"
	"../fakedriver"
	"../network"
	"time"
)

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

type Client struct {
    ID              network.ID
    LastPassedFloor int
    Timer           *time.Timer
    HasTimedOut     bool
}

func IsSameOrder(a, b Order) bool {
    return a.Button.Floor == b.Button.Floor &&
           a.Button.Type  == b.Button.Type
}

func IsNewOrder(request Order, orders []Order) bool {
    for _, o := range(orders) {
        if IsSameOrder(o, request) {
            return false
        }
    }
    return true
}

func DistributeWork(clients map[network.ID]Client, orders []Order) {
	// Distribute to closest lift
    for i, o := range(orders) {
        if (o.Button.Type != driver.ButtonOut) &&
           (o.TakenBy == network.InvalidID ||
            clients[o.TakenBy].HasTimedOut) {

            closest := closestActiveLift(clients, o.Button.Floor)
            if closest == network.InvalidID {
                log.Fatal("Cannot distribute work when there are no lifts!")
            }
            o.TakenBy = closest
            orders[i] = o
        }
    }

    for id, c := range(clients) {
        target_floor := -1
        current_pri  := -1
        for index, order := range(orders) {
            if order.TakenBy == id && order.Priority {
                target_floor = order.Button.Floor
                current_pri = index
            }
        }

        better_pri := -1
        if target_floor >= 0 {
            better_pri = closestOrderAlong(id, orders, c.LastPassedFloor, target_floor)
        } else {
            better_pri = closestOrderNear(id, orders, c.LastPassedFloor)
        }

        if better_pri >= 0 {
            if current_pri >= 0 {
                orders[current_pri].Priority = false
            }
            orders[better_pri].Priority = true
        }
    }
}

func distanceSqrd(a, b int) int {
    return (a - b) * (a - b)
}

func closestActiveLift(clients map[network.ID]Client, floor int) network.ID {
    closest_df := 100
    closest_id := network.InvalidID
    for id, client := range(clients) {
        if client.HasTimedOut {
            continue
        }
        df := distanceSqrd(client.LastPassedFloor, floor)
        if df < closest_df {
            closest_df = df
            closest_id = id
        }
    }
    return closest_id
}

func closestOrderNear(owner network.ID, orders []Order, floor int) int {
    closest_i := -1
    closest_d := -1
    for i, o := range(orders) {
        if o.TakenBy != owner {
            continue
        }
        d := distanceSqrd(o.Button.Floor, floor)
        if closest_i == -1 || d < closest_d {
            closest_i = i
            closest_d = d
        }
    }
    return closest_i
}

func closestOrderAlong(owner network.ID, orders []Order, from, to int) int {
    closest_i := -1
    closest_d := -1
    for i, o := range(orders) {
        if o.TakenBy != owner {
            continue
        }
        // Deliberately not using o.Floor >= from, since
        // the lift might not actually be at its last passed
        // floor by the time we distribute work.
        in_range   := o.Button.Floor > from && o.Button.Floor <= to
        dir_up     := to - from > 0 // Likewise, these are not using = since we
        dir_down   := to - from < 0 // assert that LPF != TF when calling this
        order_up   := o.Button.Type == driver.ButtonUp
        order_down := o.Button.Type == driver.ButtonDown
        order_out  := o.Button.Type == driver.ButtonOut
        if in_range && ((dir_up   && (order_up   || order_out)) ||
                        (dir_down && (order_down || order_out))) {
            d := distanceSqrd(o.Button.Floor, from)
            if closest_i == -1 || d < closest_d {
                closest_i = i
                closest_d = d
            }
        }
    }
    return closest_i
}
