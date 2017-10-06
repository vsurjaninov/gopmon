package main

import "fmt"

type ProcEventsListener struct{}


func (l *ProcEventsListener) onAck(event EventAck) {
    fmt.Printf("%T no=%d\n", event, event.No)
}

func (l *ProcEventsListener) onFork(event EventFork) {
    fmt.Printf("%T ppid=%d ptid=%d cpid=%d ctid=%d\n",
        event, event.ParentPid, event.ParentTid, event.ChildPid, event.ChildTid)
}

func (l *ProcEventsListener) onExec(event EventExec) {
    fmt.Printf("%T pid=%d tid=%d\n",
        event, event.Pid, event.Tid)
}

func (l *ProcEventsListener) onUid(event EventUid) {
    fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
        event, event.Pid, event.Tid, event.Ruid, event.Euid)
}

func (l *ProcEventsListener) onGid(event EventGid) {
    fmt.Printf("%T pid=%d tid=%d ruid=%d euid=%d\n",
        event, event.Pid, event.Tid, event.Rgid, event.Egid)
}

func (l *ProcEventsListener) onSid(event EventSid) {
    fmt.Printf("%T pid=%d tid=%d\n",
        event, event.Pid, event.Tid)
}

func (l *ProcEventsListener) onPtrace(event EventPtrace) {
    fmt.Printf("%T pid=%d tid=%d tpid=%d ttid=%d\n",
        event, event.TargetPid, event.TargetTid, event.TracerPid, event.TracerTid)
}

func (l ProcEventsListener) onComm(event EventComm) {
    fmt.Printf("%T pid=%d tid=%d\n",
        event, event.Pid, event.Tid)
}

func (l *ProcEventsListener) onCoreDump(event EventCoreDump) {
    fmt.Printf("%T pid=%d tid=%d\n",
        event, event.Pid, event.Tid)
}

func (l *ProcEventsListener) onExit(event EventExit) {
    fmt.Printf("%T pid=%d tid=%d code=%d signal=%d\n",
        event, event.Pid, event.Tid, event.Code, event.Signal)
}
