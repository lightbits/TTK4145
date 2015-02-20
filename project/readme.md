Todo
----

*   Design issues
    - Do we have a single backup for primary or many?
    - Do we break up large client/master updates into several packets?
    - How do we handle case when backup loses connection to primary,
    but other clients remain connected to primary (which gives two primaries)?

    - What should the network module interface be like?
        - Send/RecvByteBlob?
        - Send/RecvClientUpdate, Send/RecvMasterUpdate?
        - Send/RecvMessage where Message is an interface that has defined
        a ConvertToByteBlob() function?

    - If the elevator is disconnected from the network, it should still serve whatever orders are currently "in the system" (ie whatever lights are showing)
        - It should also serve the internal orders, so that people can exit the elevator even if it is disconnected

*   Implement timeouts
*   Implement network module
*   Implement driver
    - Interchangeable with fake/test io

*   Implement client
    - Send client updates to primary
    - Receive master updates
        - Set lights, etc
    - Give jobs to lift
    - Take over on primary timeout
    - How the elevator behaves when it cannot connect to the network during initialization
    - How the external (call up, call down) buttons work when the elevator is disconnected from the network

*   Implement master
    - Receive client updates

*   Implement job prioritization
    - Prioritize among lifts by distributing to closest?
    - Mark single job as "DoThisNow"

*   Implement lift controller
    - How do we send jobs to the lift?
    - Define inputs and outputs. Perhaps GoToFloor (in) and ClearedOrder (out).
