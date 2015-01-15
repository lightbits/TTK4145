3. Why concurrency?
===================

We like concurrent execution for a couple of reasons
-------------------------
*   Speed:
    It is easier to make processing units that have more, but less powerful, cores rather than a single, very powerful, core. Utilizing the full capabilities of your processor requires programs to be designed in such a way that they perform tasks in parallel.

    ![](test.png)
    
    Example: Fragment shader programs in GPUs are small programs that are executed at the same time for each pixel being output to the screen. They allow you to make cool effects that run really fast!

*   Multitasking: 
    Allow continued use of one application while others are running in the background. 
    Example: Watching a youtube video while writing a report (not recommended).

*   Responsiveness: 
    Allow applications to continue taking input while performing some work. 
    Example: A video game can load resources while the player moves around and does stuff.

*   Simpler translation from idea -> code:
    In traditional languages, when you call a function it will run till completion. Sometimes, this is not what you want. Instead you might want to start doing a task, and do other stuff until the task completes.

    Example: Games sometimes use loading screens to load resources. But a static screen that does nothing is boring, and the user might think that the game has crashed. So instead they usually present a swirly loading icon, some fading action scenes and unhelpful tips for how to best blast your opponents into space, to remind the user that the game is indeed still working.

    Clearly, we want to be able to render the loading screen at the same time as we load content, which is a task for concurrent programming.


But it has some drawbacks
-------------------------
* It is harder to test concurrent code, due to the apparent non-determinism of your program. In the worst case, you first detect a bug after deploying your application to the end-user.

* Concurrent tasks must have some form of synchronization, which will make your program more complex.

Terms
-------------------------
**A process** is a collection of things required to run a program. Common things include virtual memory map, file handles, window handles, process ID, etc.

**A thread** is an object that executes lines code, that can be scheduled by the OS.

**A green thread** is similiar, but the scheduling is done by the application creating the thread, instead of the OS.

**A coroutine** is a function that can be paused at any point in execution and resumed where it left off. The difference between a coroutine and a thread, is that a coroutine explicitly returns control to the main scheduler itself. For example

    coroutine FadeAnimation()
    {
        for f = 1.0 ; f >= 0.0; f -= 0.1
        {
            sprite.color.alpha = f
            yield WaitSeconds(0.1)
        }
    }

    UpdateAndRenderGame()
    {
        if isKeyPressed('f') 
        {
            FadeAnimation()
        }

        DrawRectangle(sprite.position, 
                      sprite.texture, 
                      sprite.color)
    }

Can be used as a clever way of optimizing code that only needs to be run once in a while.

    CollisionDetect()
    {
        for i = 0; i < enemies.count; i++
        {
            if Distance(enemies[i].position, player.position) < dangerDistance
            {
                return true;
            }
        }
        return false;
    }

    coroutine DoCollisionDetect()
    {
        while (1)
        {
            CollisionDetect
            yield WaitSeconds(0.1)
        }
    }

C: pthread_create()
-------------------------
This creates a native thread (that is, a thread that is scheduled by the OS and that can run on a different core).

Python: threading.Thread()
-------------------------
This creates a green thread. Python does support native OS threads, but the global interpreter lock (GIL) prevents this. Each thread that wants to run some Python code must wait for the GIL to be released by the other thread using it. Therefore, if you want to perform some heavy python computation in one thread, and still run other code, python threads will not help. The other threads will block while waiting to acquire the GIL.

The exception to this is in C modules. C extensions are .c files that you can import as modules in your Python program. They define macros that release or regain control of the GIL, respectively. For example:

    Py_BEGIN_ALLOW_THREADS
    ... some blocking operation ...
    Py_END_ALLOW_THREADS

will release the GIL before performing the blocking operation, to allow other Python code to run while performing it, and retake it afterwards.

Go: goroutine
-------------------------
Threading in Go is done through goroutines. These are multiplexed/timesliced onto OS threads as required. They are not necessarily native threads, but they are not as restrictive as Pythons threading (with GIL). If a goroutine blocks, other goroutines can keep running.

Setting GOMAXPROCS > 1 allows Go's scheduler to use more than one OS thread (and then perhaps more than one CPU).

[1]: http://www.drdobbs.com/open-source/concurrency-and-python/206103078?pgno=2
[2]: http://golang.org/doc/faq#Why_GOMAXPROCS
[3]: http://golang.org/doc/faq#Concurrency
[4]: http://jessenoller.com/blog/2009/02/01/python-threads-and-the-global-interpreter-lock
[5]: http://stackoverflow.com/questions/1739614/what-is-the-difference-between-gos-multithreading-and-pthread-or-java-threads
[6]: http://concur.rspace.googlecode.com/hg/talk/concur.html#slide-30
[7]: http://docs.unity3d.com/Manual/Coroutines.html