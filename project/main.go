package main

import (
    "fmt"
    "time"
)

type Fetcher interface {
    // Fetch returns the body of URL and
    // a slice of URLs found on that page.

    Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, ch chan bool) {

    body, urls, err := fetcher.Fetch(url)
    if depth <= 0 {
        ch <- true
        return
    }

    if err != nil {
        fmt.Println(err)
        ch <- true
        return
    }

    fmt.Printf("found: %s %q\n", url, body)

    // Fetch URLs in parallel
    subch := make(chan bool)
    for _, suburl := range urls {
        go Crawl(suburl, depth - 1, fetcher, subch)
    }

    // Synchronize
    for i := 0; i < len(urls); i++ {
        <- subch
    }

    ch <- true

    // TODO: Fetch URLs in parallel.
    // TODO: Don't fetch the same URL twice.
    // This implementation doesn't do either:
    // if depth <= 0 {
    //     return
    // }
    // body, urls, err := fetcher.Fetch(url)
    // if err != nil {
    //     fmt.Println(err)
    //     return
    // }
    // fmt.Printf("found: %s %q\n", url, body)
    // for _, u := range urls {
    //     Crawl(u, depth-1, fetcher)
    // }
    // return
}

func main() {
    ch := make(chan bool)
    go Crawl("http://golang.org/", 4, fetcher, ch)
    <-ch
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
    body string
    urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
    if res, ok := f[url]; ok {
        time.Sleep(400 * time.Millisecond)
        return res.body, res.urls, nil
    }
    return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
    "http://golang.org/": &fakeResult{
        "The Go Programming Language",
        []string{
            "http://golang.org/pkg/",
            "http://golang.org/cmd/",
        },
    },
    "http://golang.org/pkg/": &fakeResult{
        "Packages",
        []string{
            "http://golang.org/",
            "http://golang.org/cmd/",
            "http://golang.org/pkg/fmt/",
            "http://golang.org/pkg/os/",
        },
    },
    "http://golang.org/pkg/fmt/": &fakeResult{
        "Package fmt",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
    "http://golang.org/pkg/os/": &fakeResult{
        "Package os",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
}


// func Fibonacci(ch, quit chan int) {
//     x, y := 0, 1
//     for {
//         select {
//         case ch <- x:
//             x, y = y, x + y
//         case <-quit:
//             fmt.Println("quit")
//             return
//         default:
//             fmt.Println(".")
//             time.Sleep(100 * time.Millisecond)
//         }
//     }
// }

// func main() {
//     ch := make(chan int)
//     quit := make(chan int)

//     go func() {
//         for i := 0; i < 10; i++ {
//             fmt.Println(<-ch)
//             time.Sleep(1 * time.Second)
//         }
//         quit <- 0
//     }()
//     Fibonacci(ch, quit)
// }