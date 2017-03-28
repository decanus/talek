package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir"
)

import "testing"

func fromEnvOrDefault(envKey string, default_val int) int {
	if os.Getenv(envKey) != "" {
		val, _ := strconv.Atoi(os.Getenv(envKey))
		return val
	}
	return default_val
}

func testConf() Config {
	return Config{
		CommonConfig: &common.CommonConfig{
			NumBuckets:         uint64(fromEnvOrDefault("NUM_BUCKETS", 512)),
			BucketDepth:        fromEnvOrDefault("BUCKET_DEPTH", 4),
			DataSize:           fromEnvOrDefault("DATA_SIZE", 512),
			BloomFalsePositive: 0.95,
			MaxLoadFactor:      0.95,
			LoadFactorStep:     0.02,
		},
		ReadBatch:        fromEnvOrDefault("BATCH_SIZE", 8),
		WriteInterval:    time.Second,
		ReadInterval:     time.Second,
		TrustDomain:      nil,
		TrustDomainIndex: 0,
		ServerAddrs:      nil,
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

	writeReplyChan := make(chan *common.WriteReply)
	shard.Write(&common.WriteArgs{Bucket1: 0, Bucket2: 1, Data: bytes.NewBufferString("Magic").Bytes(), InterestVector: []byte{}, ReplyChan: writeReplyChan})

	// Force DB write.
	shard.syncChan <- 1

	replychan := make(chan *common.BatchReadReply)

	rv := make([]byte, 512)
	rv[0] = 0xff
	req := common.PirArgs{RequestVector: rv}
	reqs := make([]common.PirArgs, 8)
	for i := 0; i < 8; i++ {
		reqs[i] = req
	}
	shard.BatchRead(&DecodedBatchReadRequest{Args: reqs, ReplyChan: replychan})

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
	stdWrite := &common.WriteArgs{Bucket1: 0, Bucket2: 1, Data: bytes.NewBufferString("Magic").Bytes(), InterestVector: []byte{}}

	//A default read request
	reqs := make([]common.PirArgs, conf.ReadBatch)
	rv := make([]byte, int(conf.NumBuckets))
	for i := 0; i < len(rv); i++ {
		rv[i] = byte(rand.Int())
	}
	req := common.PirArgs{RequestVector: rv}
	for i := 0; i < conf.ReadBatch; i++ {
		reqs[i] = req
	}
	stdRead := &DecodedBatchReadRequest{reqs, replychan}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%readsPerWrite == 0 {
			stdWrite.Bucket1 = uint64(rand.Int()) % conf.NumBuckets
			stdWrite.Bucket2 = uint64(rand.Int()) % conf.NumBuckets
			shard.Write(stdWrite)
		} else {
			shard.BatchRead(stdRead)
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
