package libtalek

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
)

// Client represents a connection to the Talek system. Typically created with
// NewClient, the object manages requests, both reads an writes.
type Client struct {
	log    *common.Logger
	name   string
	config atomic.Value //ClientConfig
	dead   int32
	rand   *drbg.HashDrbg
	leader common.LeaderInterface

	subscriptions     []Subscription
	pendingWrites     chan *common.WriteArgs
	pendingReads      chan request
	subscriptionMutex sync.Mutex

	lastSeqNo uint64
}

type request struct {
	*common.ReadArgs
	*Subscription
}

// NewClient creates a Talek client for reading and writing metadata-protected messages.
func NewClient(name string, config ClientConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.config.Store(config)
	c.leader = leader

	rand, err := drbg.NewHashDrbg(nil)
	if err != nil {
		c.log.Error.Fatalf("Error creating Hashdrbg: %v\n", err)
		return nil
	}
	c.rand = rand

	//todo: should channel capacity be smarter?
	c.pendingReads = make(chan request, 5)
	c.pendingWrites = make(chan *common.WriteArgs, 5)

	go c.readPeriodic()
	go c.writePeriodic()

	c.log.Info.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/

func (c *Client) SetConfig(config ClientConfig) {
	c.config.Store(config)
}

func (c *Client) Kill() {
	atomic.StoreInt32(&c.dead, 1)
}

func (c *Client) Ping() bool {
	var reply common.PingReply
	err := c.leader.Ping(&common.PingArgs{Msg: "PING"}, &reply)
	if err == nil && reply.Err == "" {
		c.log.Info.Printf("Ping success\n")
		return true
	}
	c.log.Warn.Printf("Ping fail: err=%v, reply=%v\n", err, reply)
	return false
}

// MaxLength returns the maximum allowed message the client can Publish.
// TODO: support messages spanning multiple data items.
func (c *Client) MaxLength() int {
	config := c.config.Load().(ClientConfig)
	return config.DataSize
}

// Publish a new message to the end of a topic.
func (c *Client) Publish(handle *Topic, data []byte) error {
	config := c.config.Load().(ClientConfig)

	if len(data) > config.DataSize-PublishingOverhead {
		return errors.New("message too long")
	} else if len(data) < config.DataSize-PublishingOverhead {
		allocation := make([]byte, config.DataSize-PublishingOverhead)
		copy(allocation[:], data)
		data = allocation
	}

	writeArgs, err := handle.GeneratePublish(config.CommonConfig, data)
	c.log.Info.Printf("Wrote %v(%d) to %d,%d.", writeArgs.Data[0:4], len(writeArgs.Data), writeArgs.Bucket1, writeArgs.Bucket2)
	if err != nil {
		return err
	}

	c.pendingWrites <- writeArgs
	return nil
}

// Poll Subscribes to updates on a given log Subscription.
// When done reading message,s the the channel can be closed via the Done
// method.
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

// Done unsubscribes a Subscription from being Polled for new items.
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

/** Private methods **/
func (c *Client) writePeriodic() {
	var req *common.WriteArgs

	for atomic.LoadInt32(&c.dead) == 0 {
		reply := common.WriteReply{}
		conf := c.config.Load().(ClientConfig)
		select {
		case req = <-c.pendingWrites:
			break
		default:
			req = c.generateRandomWrite(conf)
		}
		err := c.leader.Write(req, &reply)
		if err != nil {
			reply.Err = err.Error()
		}
		if reply.GlobalSeqNo > c.lastSeqNo {
			c.lastSeqNo = reply.GlobalSeqNo
		}
		if req.ReplyChan != nil {
			req.ReplyChan <- &reply
		}
		//TODO: switch to poisson
		time.Sleep(conf.WriteInterval)
	}
}

func (c *Client) readPeriodic() {
	var req request

	for atomic.LoadInt32(&c.dead) == 0 {
		reply := common.ReadReply{}
		conf := c.config.Load().(ClientConfig)
		select {
		case req = <-c.pendingReads:
			break
		default:
			req = c.nextRequest(&conf)
		}
		encreq, err := req.ReadArgs.Encode(conf.TrustDomains)
		if err != nil {
			reply.Err = err.Error()
		} else {
			err := c.leader.Read(&encreq, &reply)
			if err != nil {
				reply.Err = err.Error()
			}
		}
		if reply.GlobalSeqNo.End > c.lastSeqNo {
			c.lastSeqNo = reply.GlobalSeqNo.End
		}
		if req.Subscription != nil {
			req.Subscription.OnResponse(req.ReadArgs, &reply, uint(conf.DataSize))
		}
		time.Sleep(conf.ReadInterval)
	}
}

func (c *Client) generateRandomWrite(config ClientConfig) *common.WriteArgs {
	args := &common.WriteArgs{}
	args.Bucket1 = c.rand.RandomUint64() % config.CommonConfig.NumBuckets
	args.Bucket2 = c.rand.RandomUint64() % config.CommonConfig.NumBuckets
	args.Data = make([]byte, config.CommonConfig.DataSize, config.CommonConfig.DataSize)
	c.rand.FillBytes(args.Data)
	return args
}

func (c *Client) generateRandomRead(config *ClientConfig) *common.ReadArgs {
	args := &common.ReadArgs{}
	vectorSize := uint32((config.CommonConfig.NumBuckets+7)/8 + 1)
	args.TD = make([]common.PirArgs, len(config.TrustDomains), len(config.TrustDomains))
	for i := 0; i < len(args.TD); i++ {
		args.TD[i].RequestVector = make([]byte, vectorSize, vectorSize)
		c.rand.FillBytes(args.TD[i].RequestVector)
		seed, err := drbg.NewSeed()
		if err != nil {
			c.log.Error.Fatalf("Error creating random seed: %v\n", err)
		}
		args.TD[i].PadSeed, err = seed.MarshalBinary()
		if err != nil {
			c.log.Error.Fatalf("Failed to marshal seed: %v\n", err)
		}
	}
	return args
}

func (c *Client) nextRequest(config *ClientConfig) request {
	c.subscriptionMutex.Lock()

	if len(c.subscriptions) > 0 {
		nextTopic := c.subscriptions[0]
		c.subscriptions = c.subscriptions[1:]
		c.subscriptions = append(c.subscriptions, nextTopic)

		ra1, ra2, err := nextTopic.generatePoll(config, c.lastSeqNo)
		if err != nil {
			c.subscriptionMutex.Unlock()
			c.log.Error.Fatal(err)
			return request{c.generateRandomRead(config), nil}
		}
		c.pendingReads <- request{ra2, &nextTopic}
		c.subscriptionMutex.Unlock()
		return request{ra1, &nextTopic}
	}
	c.subscriptionMutex.Unlock()

	return request{c.generateRandomRead(config), nil}
}
