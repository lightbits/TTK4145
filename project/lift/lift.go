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
