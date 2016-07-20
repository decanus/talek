package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const defaultReadInterval = int64(time.Second)
const defaultWriteInterval = int64(time.Second)

type RequestManager struct {
	log           *log.Logger
	dead          int32
	dataSize      int
	readInterval  int64
	writeInterval int64
}

func NewRequestManager(dataSize int) *RequestManager {
	rm := &RequestManager{}
	rm.log = log.New(os.Stdout, "[RequestManager:"+strconv.Itoa(dataSize)+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	rm.dead = 0
	rm.dataSize = dataSize
	rm.readInterval = defaultReadInterval
	rm.writeInterval = defaultWriteInterval

	rm.log.Printf("NewRequestManager for size=%d\n", dataSize)
	//go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

func (rm *RequestManager) SetReadInterval(period time.Duration) {
	atomic.StoreInt64(&rm.readInterval, int64(period))
}

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) readPeriodic() {
	for rm.isDead() == false {
		rm.log.Println("readPeriodic: ")
		time.Sleep(time.Duration(atomic.LoadInt64(&rm.readInterval)))
	}
}

func (rm *RequestManager) generateRandomAppend(args *common.AppendArgs) {

}

func (rm *RequestManager) writePeriodic() {
	for rm.isDead() == false {
		rm.log.Println("writePeriodic: Dummy request")
		args := &common.AppendArgs{}
		rm.generateRandomAppend(args)

		time.Sleep(time.Duration(atomic.LoadInt64(&rm.writeInterval)))
	}
}
