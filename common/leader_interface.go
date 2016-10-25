package common

type LeaderInterface interface {
	GetName() string
	Ping(args *PingArgs, reply *PingReply) error
	Write(args *WriteArgs, reply *WriteReply) error
	Read(args *ReadArgs, reply *ReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}