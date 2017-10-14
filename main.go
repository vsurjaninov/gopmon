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

	conn := &pmon.ProcListener{}
	err = conn.Connect()

	if err != nil {
		fmt.Println("error on connect: ", err)
		return
	}

	defer func() {
		conn.Close()
	}()

	go conn.ListenEvents()

	for {
		select {
		case sig := <-sigchan:
			for _, s := range signals {
				if sig == s {
					fmt.Println("shutdown")
					return
				}
			}

		case err := <-conn.Error:
			fmt.Println("error on receive: ", err)
			return
		case event := <-conn.EventAck:
			fmt.Printf("%T no=%d\n", event, event.No)
		case event := <-conn.EventFork:
			fmt.Printf("%T ppid=%d ptid=%d cpid=%d ctid=%d\n",
				event, event.ParentPid, event.ParentTid, event.ChildPid, event.ChildTid)
		case event := <-conn.EventExec:
			fmt.Printf("%T pid=%d tid=%d\n", event, event.Pid, event.Tid)
		case event := <-conn.EventUid:
			fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
				event, event.Pid, event.Tid, event.Ruid, event.Euid)
		case event := <-conn.EventGid:
			fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
				event, event.Pid, event.Tid, event.Rgid, event.Egid)
		case event := <-conn.EventSid:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-conn.EventPtrace:
			fmt.Printf("%T pid=%d tid=%d tpid=%d ttid=%d\n",
				event, event.TargetPid, event.TargetTid, event.TracerPid, event.TracerTid)
		case event := <-conn.EventComm:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-conn.EventCoreDump:
			fmt.Printf("%T pid=%d tid=%d\n",
				event, event.Pid, event.Tid)
		case event := <-conn.EventExit:
			fmt.Printf("%T pid=%d tid=%d code=%d signal=%d\n",
				event, event.Pid, event.Tid, event.Code, event.Signal)
		}
	}
}
