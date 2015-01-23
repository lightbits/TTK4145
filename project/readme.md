Communicate with lift
---------------------
The lift interface is a collection of C files. Since we're programming in Go, we need to export the relevant functions to Go. It turns out that this is super easy. Go lets you specify flags to the C compiler in your source files, which means that we can specify that we wish to statically link with the lift library.

Simply write

    // #cgo LDFLAGS: -l ./driver/lift_driver.a
    // #include <elev.h>
    import "C"

in your Go source code, and compile as usual. The functions declared in lift.h are now available for calling in Go as

    C.enum_elev_motor_direction_t dir = C.DIRN_DOWN;
    C.elev_set_motor_direction(dir)

Design
---------------------
Use an event queue of messages to send to each elevator in the network.

                    Acknowledged
Msg             |   E #1     E. #2     E. #3
Reached floor 2 |   Yes      Yes       Yes
Opened door     |   Yes      Yes       No

If a message could not be sent to an elevator, try resending the message
a couple of times. When the elevator comes back, flush event queue.

Might be messy. What if a new elevator shows up, do we send all messages
that have ever been generated? Errhgh

--

Instead: Each elevator sends its world and synchronize their view
on the world (use timestamps to see which state is newest?).

An elevator's world encompasses its own state and the state of all elevators 
it knows about.