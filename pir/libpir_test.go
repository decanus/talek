package pir

import "errors"
import "fmt"
import "math/rand"
import "os"
import "strconv"
import "testing"

func getSocket() string {
	if os.Getenv("PIR_SOCKET") != "" {
		fmt.Printf("Testing against running pird at %s.\n", os.Getenv("PIR_SOCKET"))
		return os.Getenv("PIR_SOCKET")
	}
	return fmt.Sprintf("pirtest%d.socket", rand.Int())
}

func TestConnnect(t *testing.T) {
	sockName := getSocket()
	status := make(chan int)
	go CreateMockServer(status, sockName)
	<-status

	pirServer, err := Connect(sockName)
	if err != nil {
		t.Error(err)
		return
	}

	pirServer.Disconnect()

	status <- 1
	<-status
}

func TestPir(t *testing.T) {
	sockName := getSocket()
	status := make(chan int)
	go CreateMockServer(status, sockName)
	<-status

	pirServer, err := Connect(sockName)
	if err != nil {
		t.Error(err)
		return
	}

	pirServer.Configure(512, 512, 8)
	db, err := pirServer.GetDB()
	if err != nil {
		t.Error(err)
		return
	}
	for x := range db.DB {
		db.DB[x] = byte(x)
	}

	pirServer.SetDB(db)

	responseChan := make(chan []byte)
	masks := make([]byte, 512)
	masks[0] = 0x01

	err = pirServer.Read(masks, responseChan)

	if err != nil {
		t.Error(err)
		return
	}

	response := <-responseChan
	if response == nil {
		t.Error(errors.New("no response received"))
		return
	}

	if response[1] != byte(1) {
		t.Errorf("response is incorrect. byte 1 was %d, not '1'", response[1])
	}

	pirServer.Disconnect()

	status <- 1
	<-status
}

func BenchmarkPir(b *testing.B) {
	cellLength := 1024
	cellCount := 2048
	batchSize := 8
	if os.Getenv("PIR_CELL_LENGTH") != "" {
		cellLength, _ = strconv.Atoi(os.Getenv("PIR_CELL_LENGTH"))
	}
	if os.Getenv("PIR_CELL_COUNT") != "" {
		cellCount, _ = strconv.Atoi(os.Getenv("PIR_CELL_COUNT"))
	}
	if os.Getenv("PIR_BATCH_SIZE") != "" {
		batchSize, _ = strconv.Atoi(os.Getenv("PIR_BATCH_SIZE"))
	}

	sockName := getSocket()
	status := make(chan int)
	go CreateMockServer(status, sockName)
	<-status

	pirServer, err := Connect(sockName)
	if err != nil {
		b.Error(err)
		return
	}

	pirServer.Configure(cellLength, cellCount, batchSize)
	db, err := pirServer.GetDB()
	if err != nil {
		b.Error(err)
		return
	}
	for x := range db.DB {
		db.DB[x] = byte(x)
	}

	pirServer.SetDB(db)

	responseChan := make(chan []byte)
	masks := make([]byte, cellCount*batchSize/8)
	for i := 0; i < len(masks); i++ {
		masks[i] = byte(rand.Int())
	}

	b.ResetTimer()

	signalChan := make(chan int)
	go func() {
		for j := 0; j < b.N; j++ {
			response := <-responseChan
			b.SetBytes(int64(len(response)))
		}
		signalChan <- 1
	}()

	for i := 0; i < b.N; i++ {
		err := pirServer.Read(masks, responseChan)

		if err != nil {
			b.Error(err)
		}
	}

	<-signalChan

	pirServer.Disconnect()

	status <- 1
	<-status
}
