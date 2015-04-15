Todo
----

* Design issues
    - How do we handle case when backup loses connection to primary,
    but other clients remain connected to primary (which gives two primaries)?

* Implement lift controller

Done
----
* Implement job prioritization (DistributeWork)
* Synchronize order queue across master and client
* Implement wait for backup
* Implement timeouts
* Implement network module
* Implement driver

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

For example, consider the scenario where a client sends request A.
The server acknowledges this by sending A in the next update. The
client gets this, and sets the button lamp. At this point it is
critical that A is not lost.

However, it could be the case that the master died right after it
sent the update AND the update did NOT reach the backup for some
reason. But that would be two simultaneous errors, which we assume
is very unlikely.

**No network partitioning**
