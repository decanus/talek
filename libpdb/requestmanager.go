package libpdb

import (
	//"github.com/Yawning/obfs4/common/drbg"
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"sync/atomic"
	"time"
)

//const defaultReadInterval = int64(time.Second)
//const defaultWriteInterval = int64(time.Second)

type RequestManager struct {
	log       *log.Logger
	serverRef *common.TrustDomainRef
	// Protected by `atomic`
	globalConfig *atomic.Value //*common.GlobalConfig
	dead         int32
}

func NewRequestManager(name string, serverRef *common.TrustDomainRef, globalConfig *atomic.Value) *RequestManager {
	rm := &RequestManager{}
	rm.log = log.New(os.Stdout, "[RequestManager:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	rm.serverRef = serverRef
	rm.globalConfig = globalConfig
	rm.dead = 0

	rm.log.Printf("NewRequestManager \n")
	//go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

/** PUBLIC METHODS (threadsafe) **/

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

/** PRIVATE METHODS **/
func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) writePeriodic() {
	for rm.isDead() == false {
		// Load latest config
		rm.log.Println("writePeriodic: Dummy request")
		globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
		args := &common.WriteArgs{}
		rm.generateRandomWrite(args)
		time.Sleep(globalConfig.WriteInterval)
		//time.Sleep(time.Duration(atomic.LoadInt64(&rm.writeInterval)))
	}
}

func (rm *RequestManager) readPeriodic() {
	for rm.isDead() == false {
		rm.log.Println("readPeriodic: ")
		globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
		//todo
		time.Sleep(globalConfig.ReadInterval)
		//time.Sleep(time.Duration(atomic.LoadInt64(&rm.readInterval)))
	}
}

func (rm *RequestManager) generateRandomWrite(args *common.WriteArgs) {

}
