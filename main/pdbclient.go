package main

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"log"
	"math/rand"
	"time"
)

func main() {
	log.Println("------------------")
	log.Println("--- PDB Client ---")
	log.Println("------------------")

	// Config
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "172.30.2.10:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "172.30.2.159:9000", true, false)
	trustDomainConfig2 := common.NewTrustDomainConfig("t2", "172.30.2.221:9000", true, false)
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1, trustDomainConfig2}

	leaderRpc := common.NewLeaderRpc("c0->t0", trustDomainConfig0)
	/**
	// Throughput
	numClients := 10000
	for i := 0; i < numClients; i++ {
		_ = libpdb.NewClient("c", *globalConfig, leaderRpc)
		time.Sleep(time.Duration(rand.Int()%(2*int(globalConfig.WriteInterval)/numClients)) * time.Nanosecond)
	}
	log.Printf("Generated %v clients\n", numClients)
	**/
	//c.Ping()

	// Latency
	c0 := libpdb.NewClient("c0", *globalConfig, leaderRpc)
	time.Sleep(time.Duration(rand.Int()%int(globalConfig.WriteInterval)) * time.Nanosecond)
	c1 := libpdb.NewClient("c1", *globalConfig, leaderRpc)
	startTime := time.Now()
	seqNo := c0.PublishTrace()
	for i := 0; i < 100000; i++ {
		seqNoRange := c1.PollTrace()
		if seqNoRange.Contains(seqNo) {
			log.Printf("Poll#%v: seqNo=%v in range=%v after %v\n", i, seqNo, seqNoRange, time.Since(startTime))
			break
		}
	}

	/**
	// Go on forever
	for {
		time.Sleep(10 * time.Second)
	}
	**/
}
