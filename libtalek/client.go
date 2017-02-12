package libtalek

import (
	"github.com/privacylab/talek/common"
	"sync"
	"sync/atomic"
)

/**
 * Client interface for libtalek
 * Goroutines:
 * - 1x RequestManager.writePeriodic
 * - 1x RequestManager.readPeriodic
 */
type Client struct {
	log       *common.Logger
	name      string
	config    atomic.Value //ClientConfig
	leader    common.LeaderInterface
	msgReqMan *RequestManager

	subscriptions     []Subscription
	pendingRequest    *common.ReadArgs
	pendingRequestSub RequestResponder
	subscriptionMutex sync.Mutex
}

//TODO: client needs to know the different trust domain security parameters.
func NewClient(name string, config ClientConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.config.Store(config)
	c.leader = leader

	c.msgReqMan = NewRequestManager(name, c.leader, &c.config)
	c.msgReqMan.SetReadGenerator(c)
	c.subscriptionMutex = sync.Mutex{}

	c.log.Info.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/

func (c *Client) SetConfig(config ClientConfig) {
	c.config.Store(config)
}

func (c *Client) Ping() bool {
	var reply common.PingReply
	err := c.leader.Ping(&common.PingArgs{"PING"}, &reply)
	if err == nil && reply.Err == "" {
		c.log.Info.Printf("Ping success\n")
		return true
	} else {
		c.log.Warn.Printf("Ping fail: err=%v, reply=%v\n", err, reply)
		return false
	}
}

func (c *Client) Publish(handle *Topic, data []byte) error {
	config := c.config.Load().(ClientConfig)
	write_args, err := handle.GeneratePublish(config.CommonConfig, data)
	if err != nil {
		return err
	}

	c.msgReqMan.EnqueueWrite(write_args)
	return nil
}

func (c *Client) Poll(handle *Subscription) chan []byte {
	// Check if already subscribed.
	c.subscriptionMutex.Lock()
	for x := range c.subscriptions {
		if &c.subscriptions[x] == handle {
			c.subscriptionMutex.Unlock()
			return nil
		}
	}
	c.subscriptions = append(c.subscriptions, *handle)
	c.subscriptionMutex.Unlock()

	return handle.Updates
}

func (c *Client) Done(handle *Subscription) bool {
	c.subscriptionMutex.Lock()
	for i := 0; i < len(c.subscriptions); i++ {
		if &c.subscriptions[i] == handle {
			c.subscriptions[i] = c.subscriptions[len(c.subscriptions)-1]
			c.subscriptions = c.subscriptions[:len(c.subscriptions)-1]
			c.subscriptionMutex.Unlock()
			return true
		}
	}
	c.subscriptionMutex.Unlock()
	return false
}

// Implement RequestGenerator interface for the request manager
func (c *Client) NextRequest() (*common.ReadArgs, RequestResponder) {
	config := c.config.Load().(ClientConfig)

	c.subscriptionMutex.Lock()
	if c.pendingRequest != nil {
		rec := c.pendingRequest
		rr := c.pendingRequestSub
		c.pendingRequest = nil
		c.subscriptionMutex.Unlock()
		return rec, rr
	}

	if len(c.subscriptions) > 0 {
		nextTopic := c.subscriptions[0]
		c.subscriptions = c.subscriptions[1:]
		c.subscriptions = append(c.subscriptions, nextTopic)

		ra1, ra2, err := nextTopic.generatePoll(&config, c.msgReqMan.LatestSeqNo())
		if err != nil {
			c.subscriptionMutex.Unlock()
			c.log.Error.Fatal(err)
			return nil, nil
		}
		c.pendingRequest = ra2
		c.pendingRequestSub = &nextTopic
		c.subscriptionMutex.Unlock()
		return ra1, &nextTopic
	}
	c.subscriptionMutex.Unlock()
	return nil, nil
}
