package server

import (
	"bytes"
	"fmt"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/pir"
	"math/rand"
	"os"
	"strconv"
	"time"
)

import "testing"

func fromEnvOrDefault(envKey string, default_val int) int {
	if os.Getenv(envKey) != "" {
		val, _ := strconv.Atoi(os.Getenv(envKey))
		return val
	}
	return default_val
}

func testConf() common.GlobalConfig {
	return common.GlobalConfig{
		uint64(fromEnvOrDefault("NUM_BUCKETS", 512)), // num buckets
		fromEnvOrDefault("BUCKET_DEPTH", 4),          // depth
		fromEnvOrDefault("DATA_SIZE", 512),           // data size
		fromEnvOrDefault("BATCH_SIZE", 8),            // batch size
		0.95,        // bloom false positive
		0.95,        // max load
		0.02,        // load step
		time.Second, //write interval
		time.Second, //read interval
		nil,         // trust domains
	}
}

func TestShardSanity(t *testing.T) {
	status := make(chan int)
	sock := getSocket()
	go pir.CreateMockServer(status, sock)
	<-status

	shard := NewShard("Test Shard", sock, testConf())
	if shard == nil {
		t.Error("Failed to create shard.")
		return
	}

	shard.Write(&common.WriteArgs{0, 1, bytes.NewBufferString("Magic").Bytes(), []byte{}, 0}, &common.WriteReply{})

	// Force DB write.
	shard.syncChan <- 1

	replychan := make(chan *common.BatchReadReply)
	reqs := make([]common.ReadArgs, 8)

	rv := make([]byte, 512)
	rv[0] = 0xff
	req := common.PirArgs{rv, nil}
	for i := 0; i < 8; i++ {
		reqs[i] = common.ReadArgs{[]common.PirArgs{req}}
	}
	shard.BatchRead(&common.BatchReadArgs{reqs, common.Range{0, 0, nil}, 0}, replychan)

	reply := <-replychan
	if reply.Replies[0].Data[0] != bytes.NewBufferString("Magic").Bytes()[0] {
		status <- 1
		t.Error("Failed to round-trip a write.")
		return
	}

	shard.Close()
	status <- 1
	<-status
}

func BenchmarkShard(b *testing.B) {
	fmt.Printf("Benchmark began with N=%d\n", b.N)
	readsPerWrite := fromEnvOrDefault("READS_PER_WRITE", 20)

	status := make(chan int)
	sock := getSocket()
	go pir.CreateMockServer(status, sock)
	<-status

	conf := testConf()
	shard := NewShard("Test Shard", sock, conf)
	if shard == nil {
		b.Error("Failed to create shard.")
		return
	}

	replychan := make(chan *common.BatchReadReply)

	//A default write request
	stdWrite := &common.WriteArgs{0, 1, bytes.NewBufferString("Magic").Bytes(), []byte{}, 0}

	//A default read request
	reqs := make([]common.ReadArgs, conf.ReadBatch)
	rv := make([]byte, int(conf.NumBuckets))
	for i := 0; i < len(rv); i++ {
		rv[i] = byte(rand.Int())
	}
	req := common.PirArgs{rv, nil}
	for i := 0; i < conf.ReadBatch; i++ {
		reqs[i] = common.ReadArgs{[]common.PirArgs{req}}
	}
	stdRead := &common.BatchReadArgs{reqs, common.Range{0, 0, nil}, 0}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%readsPerWrite == 0 {
			stdWrite.Bucket1 = uint64(rand.Int()) % conf.NumBuckets
			stdWrite.Bucket2 = uint64(rand.Int()) % conf.NumBuckets
			shard.Write(stdWrite, &common.WriteReply{})
		} else {
			shard.BatchRead(stdRead, replychan)
			reply := <-replychan

			if reply == nil || reply.Err != "" {
				b.Error("Read failed.")
			}
		}
		b.SetBytes(int64(1))
	}

	fmt.Printf("Benchmark called close w N=%d\n", b.N)
	shard.Close()
	status <- 1
	<-status
}
