Exercise 2: Bottlenecks
=======================

Mutex and Channel basics
------------------------
**An atomic operation** blocks other threads for the number of cycle that it takes. So it is a (possibly) multiple instruction-length operation pretending to be a single-instruction operation that cannot be interrupted.

**A semaphore** is sort of like an interrupt, and can be used like this:

    power_up() {
        draw_loading_screen()
        ... ...
        semaphoreSignal(sem_monitor_ready)
        ... ...
        semaphoreSignal(sem_power_up_done)
    }

    draw_loading_screen() {
        semaphoreWait(sem_monitor_ready)
        ... ...
        while (!semaphoreReady(sem_power_up_done)) {
            draw_loading_bar()
            ... ...
        }
    }

**A mutex** is a lock that prevents other threads from using a resource already in use. A resource could be a piece of code, like this:

    update_priority_queue(order) {
        mutex_lock()
        priority_queue.append(order)
        mutex_unlock()
    }

**A critical section** is a piece of code that accesses a shared data resource, that must not be executed by more than one thread at the same time. An example is the body of `update_priority_queue`.

Resources
=========
http://www.barrgroup.com/Embedded-Systems/How-To/RTOS-Mutex-Semaphore
http://stackoverflow.com/questions/62814/difference-between-binary-semaphore-and-mutex
https://blog.feabhas.com/2009/09/mutex-vs-semaphores-%E2%80%93-part-1-semaphores/