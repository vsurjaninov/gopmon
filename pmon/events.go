package pmon

import "fmt"

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

type EventAck struct {
	No uint32
}

func (e EventAck) String() string {
	return fmt.Sprintf("%T(no=%d)", e, e.No)
}

type EventFork struct {
	ParentTid uint32
	ParentPid uint32
	ChildPid  uint32
	ChildTid  uint32
}

func (e EventFork) String() string {
	return fmt.Sprintf("%T(ppid=%d ptid=%d cpid=%d ctid=%d)",
		e, e.ParentPid, e.ParentTid, e.ChildPid, e.ChildTid)
}

type EventExec struct {
	Tid uint32
	Pid uint32
}

func (e EventExec) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d)", e, e.Pid, e.Tid)
}

type EventUid struct {
	Tid  uint32
	Pid  uint32
	Ruid uint32
	Euid uint32
}

func (e EventUid) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d ruid=%d euid=%d)",
		e, e.Pid, e.Tid, e.Ruid, e.Euid)
}

type EventGid struct {
	Tid  uint32
	Pid  uint32
	Rgid uint32
	Egid uint32
}

func (e EventGid) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d ruid=%d euid=%d)",
		e, e.Pid, e.Tid, e.Rgid, e.Egid)
}

type EventSid struct {
	Tid uint32
	Pid uint32
}

func (e EventSid) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d)", e, e.Pid, e.Tid)
}

type EventPtrace struct {
	TargetTid uint32
	TargetPid uint32
	TracerTid uint32
	TracerPid uint32
}

func (e EventPtrace) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d tpid=%d ttid=%d)",
		e, e.TargetPid, e.TargetTid, e.TracerPid, e.TracerTid)
}

type EventComm struct {
	Tid  uint32
	Pid  uint32
	Comm [16]byte
}

func (e EventComm) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d)",
		e, e.Pid, e.Tid)
}

type EventCoreDump struct {
	Tid uint32
	Pid uint32
}

func (e EventCoreDump) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d)", e, e.Pid, e.Tid)
}

type EventExit struct {
	Tid    uint32
	Pid    uint32
	Code   uint32
	Signal uint32
}

func (e EventExit) String() string {
	return fmt.Sprintf("%T(pid=%d tid=%d code=%d signal=%d)",
		e, e.Pid, e.Tid, e.Code, e.Signal)
}
