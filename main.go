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
			fmt.Printf("%T no=%d\n", event, event.No)
		case event := <-listener.EventFork:
			fmt.Printf("%T ppid=%d ptid=%d cpid=%d ctid=%d\n",
				event, event.ParentPid, event.ParentTid, event.ChildPid, event.ChildTid)
		case event := <-listener.EventExec:
			fmt.Printf("%T pid=%d tid=%d\n", event, event.Pid, event.Tid)
		case event := <-listener.EventUid:
			fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
				event, event.Pid, event.Tid, event.Ruid, event.Euid)
		case event := <-listener.EventGid:
			fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
				event, event.Pid, event.Tid, event.Rgid, event.Egid)
		case event := <-listener.EventSid:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-listener.EventPtrace:
			fmt.Printf("%T pid=%d tid=%d tpid=%d ttid=%d\n",
				event, event.TargetPid, event.TargetTid, event.TracerPid, event.TracerTid)
		case event := <-listener.EventComm:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-listener.EventCoreDump:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-listener.EventExit:
			fmt.Printf("%T pid=%d tid=%d code=%d signal=%d\n",
				event, event.Pid, event.Tid, event.Code, event.Signal)
		}
	}
}
