// Go 1.2
// go run helloworld_go.go

package main

import (
    . "fmt"     // Using '.' to avoid prefixing functions with their package names
                // This is probably not a good idea for large projects...
    // "runtime"
    "time"
)

var i int = 0

func thread1() {
    for k := 1; k <= 1000000; k++ {
        i++

        // LOAD R1 i
        // ---------- <- interrupted here
        // INCR R1
        // PUSH R1 &i
    }
}

func thread2() {
    for k := 1; k <= 1000000; k++ {
        i--

        // LOAD R1 i
        // DECR R1
        // PUSH R1 &i
    }
}

func main() {
    // I guess this is a hint to what GOMAXPROCS does...
    // Try doing the exercise both with and without it
    // runtime.GOMAXPROCS(runtime.NumCPU())

    go thread1()
    go thread2()

    // We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
    // We'll come back to using channels in Exercise 2. For now: Sleep.
    time.Sleep(1 * time.Second)

    Println(i)
}