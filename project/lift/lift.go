/*
TODO: Move events into a seperate go file?
Common events accessible by those who need them?

Do we actually need to say that we changed direction?
What _is_ direction anyway?

...

I think the master should derive the "direction" using
more primitive variables. Such as last passed floor, and
current priority.

So, say last passed floor was A, and target floor is B.
If there is a floor C between A and B with an order, then:

    if B > A and C.order == UP:
        stop at C
    else if B < A and C.order == DOWN:
        stop at C
    else
        do not stop at C
*/

package liftcontroller

func StateMachine(InReachedFloor     chan ReachedFloorEvent,
                  OutDirectionChange chan DirectionChangeEvent) {
    for {
        select {

            case

            case e := <-Events:
                switch (e.type) {
                    case ReachedFloorEvent:
                        switch (state) {
                            case state_idle:

                        }
                }
        }
    }
}
