package main
import (
    "fmt"
    "runtime"
)

var global_i int = 0

func thread1(ch chan int, done chan int) {
    for k := 0; k < 1000; k++ {
        i := <- ch
        i++
        ch <- i
    }
    done <- 1
}

func thread2(ch chan int, done chan int) {
    for k := 0; k < 1000; k++ {
        i := <- ch
        i--
        ch <- i
    }
    done <- 1
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    i_ch := make(chan int, 1)
    done := make(chan int)
    go thread1(i_ch, done)
    go thread2(i_ch, done)
    i_ch <- global_i

    <- done
    <- done
    
    fmt.Println(<- i_ch)
}

// // Implementation 1
// func thread1(ch chan int) {
//     for i := 0; i < 1000; i++ {
//         ch <- +1
//     }
// }

// func thread2(ch chan int) {
//     for i := 0; i < 1001; i++ {
//         ch <- -1
//     }
// }

// func main() {
//     runtime.GOMAXPROCS(runtime.NumCPU())
    
//     i := 0
//     ch := make(chan int)

//     go thread1(ch)
//     go thread2(ch)

//     for k := 0; k < 2001; k++ {
//         i += <-ch
//     }

//     fmt.Println(i)
// }