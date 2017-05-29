package replica

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server/coordinator"
)

// Interface is the interface to the central coordinator
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	Notify(args *coordinator.NotifyArgs, reply *coordinator.NotifyReply) error
	Write(args *common.WriteArgs, reply *common.WriteReply) error
	Read(args *ReadArgs, reply *ReadReply) error
}
