// Go 1.2
// go run helloworld_go.go

package main

import (
    . "fmt"     // Using '.' to avoid prefixing functions with their package names
                //   This is probably not a good idea for large projects...
    "runtime"
    "time"
)

var i int = 0

func thread1() {
    for k := 1; k <= 1000000; k++ {
        i++
    }
}

func thread2() {
    for k := 1; k <= 1000000; k++ {
        i--;
    }
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())    // I guess this is a hint to what GOMAXPROCS does...
                                            // Try doing the exercise both with and without it!
    
    go thread1()
    go thread2()

    // We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
    // We'll come back to using channels in Exercise 2. For now: Sleep.
    time.Sleep(1 * time.Second)

    // This prints some ~random value between -1 mill and + 1 mill. This is because the
    // threads can be interrupted during execution to let other threads run.
    Printf("%d", i)
}