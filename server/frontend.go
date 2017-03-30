package server

import (
	"log"
	"os"

	"github.com/privacylab/talek/common"
)

// Frontend terminates client connections to the leader server.
type Frontend struct {
	// Private State
	log          *log.Logger
	name         string
	serverConfig *Config
	follower     common.FollowerInterface
	isLeader     bool

	//dataLayerRef *DataLayerRef
}

// NewFrontend creates a new Frontend for a provided configuration.
func NewFrontend(name string, serverConfig *Config, follower common.FollowerInterface, isLeader bool) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.serverConfig = serverConfig
	fe.follower = follower
	fe.isLeader = isLeader

	return fe
}

/** PUBLIC METHODS (threadsafe) **/

// Ping implements latency testing determination.
func (fe *Frontend) Ping(args *common.PingArgs, reply *common.PingReply) error {
	fe.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	if fe.follower != nil {
		var fReply common.PingReply
		fErr := fe.follower.Ping(&common.PingArgs{Msg: "PING"}, &fReply)
		if fErr != nil {
			var fName string
			fe.follower.GetName(nil, &fName)
			reply.Err = fName + " Ping failed"
		} else {
			reply.Err = fReply.Err
		}
	} else {
		reply.Err = ""
	}

	reply.Msg = "PONG"
	return nil
}

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	fe.log.Println("Write: ")
	// @TODO
	return nil
}

func (fe *Frontend) Read(args *common.EncodedReadArgs, reply *common.ReadReply) error {
	fe.log.Println("Read: ")
	// @TODO
	return nil
}

// GetUpdates provides the most recent global interest vector deltas.
func (fe *Frontend) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	fe.log.Println("GetUpdates: ")
	// @TODO
	return nil
}
