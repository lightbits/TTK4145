Todo
----

* Design issues
    - How do we handle case when backup loses connection to primary,
    but other clients remain connected to primary (which gives two primaries)?

* Implement job prioritization
    - Prioritize among lifts by distributing to closest?
    - Mark single job as "DoThisNow"

* Implement client
* Implement wait for master
* Implement master

Done
----
* Implement wait for backup
* Implement timeouts
* Implement network module
* Implement driver
* Implement lift controller

Choices made
------------
**How the elevator behaves when it cannot connect to the network during initialization:**
It will ignore any button pressing events, and thus not accept any orders,
until it hears from a master. If no master shows up, even after a long time,
an engineer should check out what is going on. Possibly restart the system.

**How the external (call up, call down) buttons work when the elevator is disconnected from the network**:
They are ignored.

**Which orders are cleared when stopping at a floor**:
All. We assume that when a lift arrives at a floor, any orders up, down or
out at that floor will be done. That is, everyone enters and/or exits when
the door opens.

Assumptions
-----------

**At least one elevator is always alive**

**Stop button & Obstruction switch are disabled**

**No multiple simultaneous errors**

**No network partitioning**
