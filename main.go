package main

import (
	"fmt"
	"github.com/vsurjaninov/gopmon/pmon"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Println("Monitoring not started. Root privileges are required!")
		return
	}

	var err error
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, signals...)

	listener := &pmon.ProcListener{}
	err = listener.Connect()

	if err != nil {
		fmt.Println("error on connect: ", err)
		return
	}

	defer listener.Close()
	go listener.ListenEvents()

	for {
		select {
		case sig := <-sigchan:
			for _, s := range signals {
				if sig == s {
					fmt.Println("shutdown")
					return
				}
			}

		case err := <-listener.Error:
			fmt.Println("error on receive: ", err)
			return
		case event := <-listener.EventAck:
			fmt.Println(event)
		case event := <-listener.EventFork:
			fmt.Println(event)
		case event := <-listener.EventExec:
			fmt.Println(event)
		case event := <-listener.EventUid:
			fmt.Println(event)
		case event := <-listener.EventGid:
			fmt.Println(event)
		case event := <-listener.EventSid:
			fmt.Println(event)
		case event := <-listener.EventPtrace:
			fmt.Println(event)
		case event := <-listener.EventComm:
			fmt.Println(event)
		case event := <-listener.EventCoreDump:
			fmt.Println(event)
		case event := <-listener.EventExit:
			fmt.Println(event)
		}
	}
}
