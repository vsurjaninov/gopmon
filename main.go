package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		msgs []syscall.NetlinkMessage
		err  error
	)

	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, signals...)

	conn := &CnConnection{}
	err = conn.Connect()

	if err != nil {
		fmt.Println("error on connect: ", err)
		return
	}

	defer func() {
		conn.Close()
	}()

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

		default:
			msgs, err = conn.Receive()
			if err != nil {
				fmt.Println("error on receive: ", err)
			}

			for i := range msgs {
				ParseProcEvent(msgs[i].Data)
			}
		}
	}
}
