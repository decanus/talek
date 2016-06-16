package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"net/rpc"
	"os"
)

type Client struct {
	log       *log.Logger
	name      string
	servers   []string
	msgReqMan *RequestManager
}

func Call(addr string, methodName string, args interface{}, reply interface{}) error {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Printf("rpc dialing failed: %v\n", err)
		return err
	}
	defer client.Close()

	err = client.Call(methodName, args, reply)
	if err != nil {
		log.Printf("rpc error:", err)
		return err
	}
	return nil
}

func NewClient(name string, servers []string) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.servers = servers
	// @todo update
	c.msgReqMan = NewRequestManager(8)
	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	c.log.Printf("Ping: enter\n")
	args := &common.PingArgs{"PING"}
	var reply common.PingReply
	Call(c.servers[0], "FrontEndRpc.Ping", args, &reply)
	c.log.Printf("Ping: %v, %v", args, reply)
	return true
}
