package pir

import (
	"fmt"
	"github.com/privacylab/talek/common"
	"github.com/willf/bitset"
)

// Shard represents a read-only shard of the database
// Databases are range partitioned by bucket.
// Thus, a shard represents a range of `numBuckets` buckets,
// where each bucket is []byte of length `bucketSize`.
// Note: len(data) must equal (numBuckets * bucketSize)
type ShardCPU struct {
	// Private State
	log         *common.Logger
	name        string
	bucketSize  int
	numBuckets  int
	data        []byte
	readVersion int
}

// Creates a new CPU-backed shard
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCPU(name string, bucketSize int, data []byte, readVersion int) (*ShardCPU, error) {
	s := &ShardCPU{}
	s.log = common.NewLogger(name)
	s.name = name

	// GetNumBuckets will compute the number of buckets stored in the Shard
	// If len(s.data) is not cleanly divisible by s.bucketSize,
	// returns an error
	if len(data)%bucketSize != 0 {
		return nil, fmt.Errorf("NewShardCPU(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
	}

	s.bucketSize = bucketSize
	s.numBuckets = (len(data) / bucketSize)
	s.data = data
	return s, nil
}

// Free currently does nothing. ShardCPU waits for the go garbage collector
func (s *ShardCPU) Free() error {
	return nil
}

func (s *ShardCPU) GetName() string {
	return s.name
}

func (s *ShardCPU) GetBucketSize() int {
	return s.bucketSize
}

func (s *ShardCPU) GetNumBuckets() int {
	return s.numBuckets
}

// Return a slice of the data
func (s *ShardCPU) GetData() []byte {
	return s.data[:]
}

// Insert copies the given byte array into the specified bucket at a given offset
// Returns the number of bytes copied. This will be <len(toCopy) if toCopy is
// Note: This function will only output a warning if it overwrites into the next bucket
/**
func (s *ShardCPU) Insert(bucket int, offset int, toCopy []byte) int {
	index := (bucket * s.bucketSize) + offset
	dst := s.data[index:]
	if len(toCopy) > (s.bucketSize - offset) {
		s.log.Warn.Printf("Shard.Insert overwriting next bucket\n")
	}
	return copy(dst, toCopy)
}
**/

// Read handles a batch read, where each request is represented by a BitSet
// Returns: a single byte array where responses are concatenated by the order in `reqs`
func (s *ShardCPU) Read(reqs []bitset.BitSet) ([]byte, error) {
	if s.readVersion == 0 {
		return s.read0(reqs)
	} else if s.readVersion == 1 {
		return s.read1(reqs)
	} else {
		return nil, fmt.Errorf("ShardCPU.Read: invalid readVersion=%d", s.readVersion)
	}
}

func (s *ShardCPU) read0(reqs []bitset.BitSet) ([]byte, error) {
	responses := make([]byte, len(reqs)*s.bucketSize)

	// calculate PIR
	for reqIndex := 0; reqIndex < len(reqs); reqIndex++ {
		for bucketIndex := 0; bucketIndex < s.numBuckets; bucketIndex++ {
			if reqs[reqIndex].Test(uint(bucketIndex)) {
				bucket := s.data[(bucketIndex * s.bucketSize):]
				response := responses[(reqIndex * s.bucketSize):]
				for offset := 0; offset < s.bucketSize; offset++ {
					response[offset] ^= bucket[offset]
				}
			}
		}
	}

	return responses, nil
}

func (s *ShardCPU) read1(reqs []bitset.BitSet) ([]byte, error) {
	responses := make([]byte, len(reqs)*s.bucketSize)

	// calculate PIR
	for bucketIndex := 0; bucketIndex < s.numBuckets; bucketIndex++ {
		for reqIndex := 0; reqIndex < len(reqs); reqIndex++ {
			if reqs[reqIndex].Test(uint(bucketIndex)) {
				bucket := s.data[(bucketIndex * s.bucketSize):]
				response := responses[(reqIndex * s.bucketSize):]
				for offset := 0; offset < s.bucketSize; offset++ {
					response[offset] ^= bucket[offset]
				}
			}
		}
	}

	return responses, nil
}
