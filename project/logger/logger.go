package logger

import (
    "fmt"
    "log"
    "strings"
)

type Level int
const (
    Debug Level = iota
    Info
    Error
    Fatal
)

const log_level Level = Info
const module_filter = ""

func Println(level Level, module string, args...interface{}) {
    if level >= log_level && (module_filter == "" || strings.Contains(module_filter, module)) {
        fmt.Printf("%s", fmt.Sprintf("%s\t+\t%s", module, fmt.Sprintln(args...)))
    }
    if level == Fatal {
        log.Fatal("Program terminated")
    }
}