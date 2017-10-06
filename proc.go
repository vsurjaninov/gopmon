package main

import (
    "encoding/binary"
    "fmt"
    "bytes"
)

const (
    PROC_EVENT_NONE     = 0x00000000
    PROC_EVENT_FORK     = 0x00000001
    PROC_EVENT_EXEC     = 0x00000002
    PROC_EVENT_UID      = 0x00000004
    PROC_EVENT_GID      = 0x00000040
    PROC_EVENT_SID      = 0x00000080
    PROC_EVENT_PTRACE   = 0x00000100
    PROC_EVENT_COMM     = 0x00000200
    PROC_EVENT_COREDUMP = 0x40000000
    PROC_EVENT_EXIT     = 0x80000000
)

type eventMsgHeader struct {
    What      uint32
    Cpu       uint32
    Timestamp uint64
}

type Listener interface {
    onAck(event EventAck)
    onFork(event EventFork)
    onExec(event EventExec)
    onUid(event EventUid)
    onGid(event EventGid)
    onSid(event EventSid)
    onPtrace(event EventPtrace)
    onComm(event EventComm)
    onCoreDump(event EventCoreDump)
    onExit(event EventExit)
}

type listenerCaller interface {
    callListener(listener *Listener)
}

type EventAck struct {
    No uint32
}

func (e EventAck) callListener(listener *Listener) {
    (*listener).onAck(e)
}

type EventFork struct {
    ParentTid uint32
    ParentPid uint32
    ChildPid  uint32
    ChildTid  uint32
}

func (e EventFork) callListener(listener *Listener) {
    (*listener).onFork(e)
}

type EventExec struct {
    Tid  uint32
    Pid  uint32
}

func (e EventExec) callListener(listener *Listener) {
    (*listener).onExec(e)
}

type EventUid struct {
    Tid  uint32
    Pid  uint32
    Ruid uint32
    Euid uint32
}

func (e EventUid) callListener(listener *Listener) {
    (*listener).onUid(e)
}

type EventGid struct {
    Tid  uint32
    Pid  uint32
    Rgid uint32
    Egid uint32
}

func (e EventGid) callListener(listener *Listener) {
    (*listener).onGid(e)
}

type EventSid struct {
    Tid uint32
    Pid uint32
}

func (e EventSid) callListener(listener *Listener) {
    (*listener).onSid(e)
}

type EventPtrace struct {
    TargetTid uint32
    TargetPid uint32
    TracerTid uint32
    TracerPid uint32
}

func (e EventPtrace) callListener(listener *Listener) {
    (*listener).onPtrace(e)
}

type EventComm struct {
    Tid  uint32
    Pid  uint32
    Comm [16]byte
}

func (e EventComm) callListener(listener *Listener) {
    (*listener).onComm(e)
}

type EventCoreDump struct {
    Tid uint32
    Pid uint32
}

func (e EventCoreDump) callListener(listener *Listener) {
    (*listener).onCoreDump(e)
}

type EventExit struct {
    Tid    uint32
    Pid    uint32
    Code   uint32
    Signal uint32
}

func (e EventExit) callListener(listener *Listener) {
    (*listener).onExit(e)
}

func HandleProcEvent(listener Listener, data []byte) {
    buf := bytes.NewBuffer(data)
    msg := &cnMsg{}
    hdr := &eventMsgHeader{}

    binary.Read(buf, binary.LittleEndian, msg)
    binary.Read(buf, binary.LittleEndian, hdr)

    var event listenerCaller

    switch hdr.What {
    case PROC_EVENT_NONE:
        event = &EventAck{}
    case PROC_EVENT_FORK:
        event = &EventFork{}
    case PROC_EVENT_EXEC:
        event = &EventExec{}
    case PROC_EVENT_UID:
        event = &EventUid{}
    case PROC_EVENT_GID:
        event = &EventGid{}
    case PROC_EVENT_SID:
        event = &EventSid{}
    case PROC_EVENT_PTRACE:
        event = &EventPtrace{}
    case PROC_EVENT_COMM:
        event = &EventComm{}
    case PROC_EVENT_COREDUMP:
        event = &EventCoreDump{}
    case PROC_EVENT_EXIT:
        event = &EventExit{}
    default:
        fmt.Printf("Unknown event type: 0x%08x\n", hdr.What)
        return
    }

    binary.Read(buf, binary.LittleEndian, event)
    fmt.Printf("Received %T at %d\n", event, hdr.Timestamp)
    event.callListener(&listener)
}
