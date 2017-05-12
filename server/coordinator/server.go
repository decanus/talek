package coordinator

import (
	"sync/atomic"

	"github.com/privacylab/talek/common"
	//"github.com/privacylab/talek/cuckoo"
	//"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log  *common.Logger
	name string

	// Thread-safe
	config    atomic.Value  // Config
	commitLog []*CommitArgs // Append and read only

	// Channels
	commitChan chan *CommitArgs
}

// NewServer creates a new Centralized talek server.
func NewServer(name string, config common.Config) *Server {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.config.Store(config)

	return s
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *GetInfoReply) error {
	reply.Err = ""
	reply.Name = s.name
	return nil
}

// GetCommonConfig returns the common global config
func (s *Server) GetCommonConfig(args *interface{}, reply *common.Config) error {
	config := s.config.Load().(common.Config)
	*reply = config
	return nil
}

// Commit accepts a single Write to commit. The
func (s *Server) Commit(args *CommitArgs, reply *CommitReply) error {
	s.commitChan <- args
	reply.Err = ""
	return nil
}

/**
func (s *Server) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
}
**/

/**********************************
 * PUBLIC LOCAL METHODS (threadsafe)
 **********************************/

// Close shuts down the server
func (s *Server) Close() {
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/

// processCommits will read from s.commitChan and properly trigger work
func (s *Server) processCommits() {
	var commit *CommitArgs
	//conf := s.config.Load().(common.Config)

	for {
		select {
		case commit = <-s.commitChan:
			s.commitLog = append(s.commitLog, commit)
		}
	}
}
