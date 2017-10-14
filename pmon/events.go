package pmon

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

type EventFork struct {
	ParentTid uint32
	ParentPid uint32
	ChildPid  uint32
	ChildTid  uint32
}

type EventExec struct {
	Tid uint32
	Pid uint32
}

type EventUid struct {
	Tid  uint32
	Pid  uint32
	Ruid uint32
	Euid uint32
}

type EventGid struct {
	Tid  uint32
	Pid  uint32
	Rgid uint32
	Egid uint32
}

type EventSid struct {
	Tid uint32
	Pid uint32
}

type EventPtrace struct {
	TargetTid uint32
	TargetPid uint32
	TracerTid uint32
	TracerPid uint32
}

type EventComm struct {
	Tid  uint32
	Pid  uint32
	Comm [16]byte
}

type EventCoreDump struct {
	Tid uint32
	Pid uint32
}

type EventExit struct {
	Tid    uint32
	Pid    uint32
	Code   uint32
	Signal uint32
}
