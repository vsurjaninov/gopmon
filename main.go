package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    var err  error
    signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
    sigchan := make(chan os.Signal, 1)
    evschan := make(chan RawMsg)

    signal.Notify(sigchan, signals...)

    listener := &ProcEventsListener{}
    conn := &CnConnection{}
    err = conn.Connect()

    if err != nil {
        fmt.Println("error on connect: ", err)
        return
    }

    defer func() {
        conn.Close()
    }()

    go conn.RecvRawProcEvents(evschan)

    for {
        select {
        case sig, ok := <-sigchan:
            if !ok {
                fmt.Println("Signal channel is close, exit")
                return
            }

            for _, s := range signals {
                if sig == s {
                    fmt.Println("shutdown")
                    return
                }
            }

        case msg := <-evschan:
            if msg.Error != nil {
                fmt.Println("error on receive: ", err)
            } else {
                HandleProcEvent(listener, msg.Data)
            }
        }
    }
}
