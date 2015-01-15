package main
import (
    "fmt"
    "runtime"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    ch := make(chan int)
    go func(ch chan int) {
        for i := 0; i < 20000; i++ {
            ch <- +1
        }
    }(ch)

    go func(ch chan int) {
        for i := 0; i < 20000; i++ {
            ch <- -1
        }
    }(ch)

    global_i := 0
    for i := 0; i < 40000; i++ {
        global_i += <-ch
    }
    fmt.Println(global_i)
}