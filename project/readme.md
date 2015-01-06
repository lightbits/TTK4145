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