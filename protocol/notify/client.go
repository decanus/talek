package notify

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
)

// Client is a stub for RPCs to servers to notify
type Client struct {
	log     *common.Logger
	name    string
	address string
	client  *rpc.Client
	lastErr error
}

// NewClient instantiates a client stub
func NewClient(name string, address string) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	c.client = nil // Lazily dial as necessary
	return c
}

// Close will close the RPC client
func (c *Client) Close() error {
	if c.client != nil {
		c.lastErr = c.client.Close()
		c.client = nil
		return c.lastErr
	}
	return nil
}

// Notify tells the server of a new snapshot
func (c *Client) Notify(args *Args, reply *Reply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Notify", &args, reply)
	return c.lastErr
}
