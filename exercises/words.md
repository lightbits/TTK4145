* Context:
    Minimal set of data necessary to be able to interrupt a thread and resume it later.
    For example, the program counter is absolutely necessary.
    Stored in processor registers and in the process memory.

* Context switch:
    Storing the state (context) of one thread and restoring the state of another.
    That is, pausing one thread to allow the other to run.
    Are pretty expensive, but can be made cheaper if the switch is between two threads of the same process. Threads of the same process share the same virtual memory.

* CPU affinity:
    Binding a thread to only execute on a designated range of CPUs instead of any CPU.

* Concurrency: 
    A way to structure your program into tasks that can be performed independently. 
    For example: Handling mouse and keyboard input such that the user can use both at the same time.

* Parallellism: 
    Executing multiple computations at the same time.
    For example: a vector dot product.

* Process:
    A computer program being executed.

* Thread:
    Multiple threads can exist within the same process, and share resources.

* Concurrency vs parallelism:
    A concurrent program can be run on a single-core machine, for example by interleaving processing time between the different tasks. On the other hand, a parallel program can _not_ be run on a single-core machine. Parallelism is one way of implementing concurrency.

* Process: 
    Provides resources to run a program. Resources are allocated by the OS, and include a virtual memory address space, environment variables, windows, executable code and at least one executing thread. Typically preemtively multitasked.

* Threads:
    A thing which can be scheduled by the OS to run code of a process. All threads of the same process share its virtual address space and system resources. Can be interrupted by the OS to allow a different thread to run. Can run in parallel with other threads on multiple processors.

* Green threads:
    A thread that lives in userspace (meaning it cannot access memory of other processes), and is not scheduled by the OS. Cannot take advantage of multiple CPUs.

* Fibers:
    OS managed threads that co-operatively multitask (meaning the threads themselves will relinquish control to let other threads run).

* Atomic:
    All but one thread's execution halts while a thread performs an atomic operation. This essentially makes an operation - which may be multiple instructions long - look like a single machine instruction. 

    This can work well on single-core machines, since it prevents other threads from doing stuff they shouldn't be doing. But it's poor for performance on multi-core machines, since all cores but one are prevented from doing work during an atomic operation!

* Locking:
    A thread's execution is halted if it tries to access data that is currently in use by a different thread. When the thread locking the data releases its lock, the thread can continue. This is different from atomics in that only threads that wish to access the data will halt.