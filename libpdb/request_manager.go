package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/drbg"
	"log"
	"os"
	"sync/atomic"
	"time"
)

//const defaultReadInterval = int64(time.Second)
//const defaultWriteInterval = int64(time.Second)

type RequestManager struct {
	log    *log.Logger
	leader common.LeaderInterface
	// Protected by `atomic`
	globalConfig *atomic.Value //*common.GlobalConfig
	dead         int32
	// Channels
	writeChan  chan *common.WriteArgs
	writeQueue []*common.WriteArgs
	readChan   chan *common.ReadArgs
	readQueue  []*common.ReadArgs
}

func NewRequestManager(name string, leader common.LeaderInterface, globalConfig *atomic.Value) *RequestManager {
	rm := &RequestManager{}
	rm.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	rm.leader = leader
	rm.globalConfig = globalConfig
	rm.dead = 0

	rm.log.Printf("NewRequestManager \n")
	go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

/** PUBLIC METHODS (threadsafe) **/

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

func (rm *RequestManager) EnqueueWrite(args *common.WriteArgs) {
	rm.writeChan <- args
}

func (rm *RequestManager) EnqueueRead(args *common.ReadArgs) {
	rm.readChan <- args
}

/** PRIVATE METHODS **/
func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) writePeriodic() {
	rand, randErr := drbg.NewHashDrbg(nil)
	if randErr != nil {
		rm.log.Fatalf("Error creating new HashDrbg: %v\n", randErr)
	}

	for rm.isDead() == false {
		select {
		case msg := <-rm.writeChan:
			rm.writeQueue = append(rm.writeQueue, msg)
		default:
			globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
			var args *common.WriteArgs
			var reply common.WriteReply
			if len(rm.writeQueue) > 0 {
				args = rm.writeQueue[0]
				rm.writeQueue = rm.writeQueue[1:]
				rm.log.Printf("writePeriodic: Real request to %v, %v \n", args.Bucket1, args.Bucket2)
			} else {
				args = &common.WriteArgs{}
				rm.generateRandomWrite(globalConfig, rand, args)
				rm.log.Printf("writePeriodic: Dummy request to %v, %v \n", args.Bucket1, args.Bucket2)
			}
			//@todo Do something with response
			err := rm.leader.Write(args, &reply)
			if err != nil || reply.Err != "" {
				rm.log.Printf("writePeriodic error: %v, reply=%v\n", err, reply)
			}
			time.Sleep(globalConfig.WriteInterval)
			//time.Sleep(time.Duration(atomic.LoadInt64(&rm.writeInterval)))
		}
	}
}

func (rm *RequestManager) readPeriodic() {
	rand, randErr := drbg.NewHashDrbg(nil)
	if randErr != nil {
		rm.log.Fatalf("Error creating new HashDrbg: %v\n", randErr)
	}

	for rm.isDead() == false {
		select {
		case msg := <-rm.readChan:
			rm.readQueue = append(rm.readQueue, msg)
		default:
			globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
			var args *common.ReadArgs
			var reply common.ReadReply
			if len(rm.readQueue) > 0 {
				args = rm.readQueue[0]
				rm.readQueue = rm.readQueue[1:]
				rm.log.Printf("readPeriodic: Real request \n")
			} else {
				args = &common.ReadArgs{}
				rm.generateRandomRead(globalConfig, rand, args)
				rm.log.Printf("readPeriodic: Dummy request \n")
			}
			//@todo Do something with response
			err := rm.leader.Read(args, &reply)
			if err != nil || reply.Err != "" {
				rm.log.Printf("readPeriodic error: %v, reply=%v\n", err, reply)
			}
			time.Sleep(globalConfig.ReadInterval)
			//time.Sleep(time.Duration(atomic.LoadInt64(&rm.readInterval)))
		}
	}
}

func (rm *RequestManager) generateRandomWrite(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.WriteArgs) {
	args.Bucket1 = rand.RandomUint64() % globalConfig.NumBuckets
	args.Bucket2 = rand.RandomUint64() % globalConfig.NumBuckets
	args.Data = make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	rand.FillBytes(args.Data)
}

func (rm *RequestManager) generateRandomRead(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.ReadArgs) {
	numTds := len(globalConfig.TrustDomains)
	numBytes := (uint32(globalConfig.NumBuckets) / uint32(8)) + 1
	if (uint32(globalConfig.NumBuckets) % uint32(8)) > 0 {
		numBytes = numBytes + 1
	}
	args.ForTd = make([]common.PIRArgs, numTds, numTds)
	for i := 0; i < numTds; i++ {
		args.ForTd[i].RequestVector = make([]byte, numBytes, numBytes)
		rand.FillBytes(args.ForTd[i].RequestVector)
		seed, seedErr := drbg.NewSeed()
		if seedErr != nil {
			rm.log.Fatalf("Error creating new Seed: %v\n", seedErr)
		}
		args.ForTd[i].PadSeed = seed.Export()
	}
	//args.RequestVector = make([]byte, numBytes, numBytes)
	//rand.FillBytes(&args.RequestVector)
	//@todo Trim last byte to expected number of bits

}
