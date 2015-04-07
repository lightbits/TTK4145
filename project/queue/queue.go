package queue

type Order struct {
    Button   driver.OrderButton
    TakenBy  network.ID
    Done     bool
    Priority bool
}

// TODO: Don't like this being defined in this package...
type Client struct {
    ID              network.ID
    LastPassedFloor int
    TargetFloor     int
    Timer           *time.Timer
    HasTimedOut     bool
}

func DistributeWork(clients map[network.ID]Client, orders []Order) {
    // Broad-phase distribution
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

    // Narrow-phase distribution (sort each lift queue)
    for id, c := range(clients) {
        // If the client is already heading towards a floor, we
        // don't want to change its direction. But we if there
        // is a new floor that is closer along the way, we can
        // stop there first. But only if that order is also
        // headed the same way...

        // Note that the LPF will eventually equal TF, as the client
        // can only go to the one floor which master marks as PRIORITY.
        if c.LastPassedFloor == c.TargetFloor {
            closest := closestOrderNear(id, orders, c.LastPassedFloor)
            orders[closest].Priority = true
        } else {
            closest := closestOrderAlong(id, orders, c.LastPassedFloor, c.TargetFloor)
            orders[closest].Priority = true
        }
    }
}

func distanceSqrd(a, b int) int {
    return (a - b) * (a - b)
}

func closestActiveLift(clients map[network.ID]Client, floor int) network.ID {
    closest_df := -1
    closest_id := network.InvalidID
    for id, client := range(clients) {
        if client.HasTimedOut {
            continue
        }
        df := distanceSqrd(client.LastPassedFloor, floor)
        if closest_df == -1 || df < closest_df {
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
