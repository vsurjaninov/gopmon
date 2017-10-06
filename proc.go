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

type procEventAck struct {
    Error uint32
}

type procEventFork struct {
    ParentTid uint32
    ParentPid uint32
    ChildPid  uint32
    ChildTgid uint32
}

type procEventExec struct {
    ProcessPid  uint32
    ProcessTgid uint32
}

type procEventUid struct {
    ProcessPid  uint32
    ProcessTgid uint32
    Ruid 		uint32
    Euid 		uint32
}

type procEventGid struct {
    ProcessPid  uint32
    ProcessTgid uint32
    Rgid 		uint32
    Egid 		uint32
}

type procEventSid struct {
    ProcessPid  uint32
    ProcessTgid uint32
}

type procEventPtrace struct {
    ProcessPid  uint32
    ProcessTgid uint32
    TracerPid   uint32
    TracerTgid  uint32
}

type procEventComm struct {
    ProcessPid  uint32
    ProcessTgid uint32
    Comm		[16]byte
}

type procEventCoreDump struct {
    ProcessPid  uint32
    ProcessTgid uint32
}

type procEventExit struct {
    ProcessPid  uint32
    ProcessTgid uint32
    ExitCode    uint32
    ExitSignal  uint32
}

func ParseProcEvent(data []byte) {
    buf := bytes.NewBuffer(data)
    msg := &cnMsg{}
    hdr := &eventMsgHeader{}

    binary.Read(buf, binary.LittleEndian, msg)
    binary.Read(buf, binary.LittleEndian, hdr)

    var event interface{}

    switch hdr.What {
    case PROC_EVENT_NONE:
        event = &procEventAck{}
    case PROC_EVENT_FORK:
        event = &procEventFork{}
    case PROC_EVENT_EXEC:
        event = &procEventExec{}
    case PROC_EVENT_UID:
        event = &procEventUid{}
    case PROC_EVENT_GID:
        event = &procEventGid{}
    case PROC_EVENT_SID:
        event = &procEventSid{}
    case PROC_EVENT_PTRACE:
        event = &procEventPtrace{}
    case PROC_EVENT_COMM:
        event = &procEventComm{}
    case PROC_EVENT_COREDUMP:
        event = &procEventCoreDump{}
    case PROC_EVENT_EXIT:
        event = &procEventExit{}
    default:
        fmt.Printf("Unknown event type: 0x%08x\n", hdr.What)
        return
    }

    binary.Read(buf, binary.LittleEndian, event)
    fmt.Printf("Received %T at %d\n", event, hdr.Timestamp)
}
