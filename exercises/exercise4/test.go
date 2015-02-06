package main 

import (
    "fmt"
)

// libraries must be saved with lib___.a
// but we only specify the name in the middle when linking!

// gcc -c awesome.c -oawesome.o
// ar rcs libawesome.a awesome.o

// #include "awesome.h"
// #cgo LDFLAGS: -L. -lawesome
/*
int
ThisFunctionIsNotIncluded(int x);
*/
import "C"

func main() {
    var x C.int = 2

    fmt.Println(C.ThisFunctionIsNotIncluded(x))
}