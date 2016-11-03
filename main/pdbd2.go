package main

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/server"
	"log"
	"time"
)

func main() {
	log.Println("------------------")
	log.Println("--- PDB Server ---")
	log.Println("------------------")

	// For trace debug status
	go http.ListenAndServe("localhost:8080", nil)

	// Config
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}


	s := server.NewCentralized("s", *globalConfig, nil, false)
	//s := server.NewCentralized("s", globalConfig, common.NewFollowerRpc("t0->t1", trustDomainConfig1), true)
	_ = server.NewNetworkRpc(s, 9001)

	log.Println(s)
  for {
	  time.Sleep(10 * time.Second)
  }

}
