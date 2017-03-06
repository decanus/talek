package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir"
	"github.com/privacylab/talek/server"
)

var configPath = flag.String("config", "../commonconfig.json", "Talek Common Configuration")
var trustDomainPath = flag.String("trust", "../demoleaderdomain.json", "Server key")
var pirSocket = flag.String("socket", "../../pird/pir.socket", "PIR daemon socket")

// Server starts a single, centralized talek server operating with an explicit configuration.
func main() {
	log.Println("--------------------")
	log.Println("--- Talek Server ---")
	log.Println("--------------------")
	flag.Parse()

	config := common.CommonConfigFromFile(*configPath)
	tdString, err := ioutil.ReadFile(*trustDomainPath)
	if err != nil {
		log.Printf("Could not read %s!\n", *trustDomainPath)
		return
	}
	td := new(common.TrustDomainConfig)
	if err := json.Unmarshal(tdString, td); err != nil {
		log.Printf("Could not parse %s: %v\n", *trustDomainPath, err)
		return
	}
	serverConfig := server.Config{
		CommonConfig:     config,
		WriteInterval:    time.Second,
		ReadInterval:     time.Second,
		ReadBatch:        8,
		TrustDomain:      td,
		TrustDomainIndex: 0,
		ServerAddrs:      nil,
	}

	mockPirStatus := make(chan int)
	usingMock := false
	if len(*pirSocket) == 0 {
		*pirSocket = fmt.Sprintf("pirtest%d.socket", rand.Int())
		go pir.CreateMockServer(mockPirStatus, *pirSocket)
		<-mockPirStatus
		usingMock = true
	}

	s := server.NewCentralized(td.Name, *pirSocket, serverConfig, nil, false)
	_, port, _ := net.SplitHostPort(td.Address)
	pnum, _ := strconv.Atoi(port)
	_ = server.NewNetworkRpc(s, pnum)

	log.Println("Running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	s.Close()
	if usingMock {
		mockPirStatus <- 1
		<-mockPirStatus
	}
}
