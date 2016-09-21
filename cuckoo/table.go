package cuckoo

import (
	"log"
	"math/rand"
	"os"
)

const MAX_EVICTIONS int = 500

type Comparable interface {
	Compare(other Comparable) int
}

type BucketLocation struct {
	Bucket1 int
	Bucket2 int
}

type Bucket struct {
	// `entries` and `filled` must be the same size
	data      []Comparable     // Stores actual data. Validity of an entry determined by `filled`
	bucketLoc []BucketLocation // Stores the 2 bucket locations for each entry
	filled    []bool           // False if cell is empty. Only read `t.entries[i]` if `t.filled[i]==true`
}

type Table struct {
	log        *log.Logger
	name       string
	numBuckets int // Number of buckets
	depth      int // Capacity of each bucket
	rand       *rand.Rand

	buckets []*Bucket // Data
}

// Creates a brand new cuckoo table
// Two cuckoo tables will have identical state iff,
// 1. the same randSeed is used
// 2. the same operations are applied in the same order
// numBuckets = number of buckets
// depth = the number of entries per bucket
// randSeed = seed for PRNG
func NewTable(name string, numBuckets int, depth int, randSeed int64) *Table {
	t := &Table{}
	t.log = log.New(os.Stdout, "[Cuckoo:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	t.name = name
	t.numBuckets = numBuckets
	t.depth = depth
	t.rand = rand.New(rand.NewSource(randSeed))

	t.buckets = make([]*Bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		t.buckets[i] = &Bucket{}
		t.buckets[i].data = make([]Comparable, depth)
		t.buckets[i].bucketLoc = make([]BucketLocation, depth)
		// We assume this will be filled with `false` as per bool's default value
		t.buckets[i].filled = make([]bool, depth)
	}
	return t
}

/********************
 * PUBLIC METHODS
 ********************/

// Checks if value exists in specified buckets
// Returns true if `value.Equals(...)` returns true
func (t *Table) Contains(bucket1 int, bucket2 int, value Comparable) bool {
	result := false
	result = result || t.isInBucket(bucket1, value)
	result = result || t.isInBucket(bucket2, value)
	return result
}

// Inserts the value into the cuckoo table
// Returns true on success, false on failure
// - fails if insertion cannot complete because reached MAX_EVICTIONS
// - failures return an evicted value and must trigger a rebuild by the caller
func (t *Table) Insert(bucket1 int, bucket2 int, value Comparable) (int, int, Comparable, bool) {
	var ok bool
	var nextBucket int
	currBucketLoc := BucketLocation{bucket1, bucket2}
	currVal := value

	// Randomly select 1 bucket first
	coin := t.rand.Int() % 2 // Coin could be -1, 0 or 1
	if coin == 0 {
		ok = t.tryInsertToBucket(bucket1, currBucketLoc, currVal)
		nextBucket = bucket2
	} else {
		ok = t.tryInsertToBucket(bucket2, currBucketLoc, currVal)
		nextBucket = bucket1
	}
	if ok {
		return -1, -1, nil, true
	}

	// Then try the other bucket, starting the eviction loop
	for i := 0; i < MAX_EVICTIONS; i++ {
		nextBucket, currBucketLoc, currVal, ok = t.insertAndEvict(nextBucket, currBucketLoc, currVal)
		if ok {
			t.log.Printf("Insert: %v evictions\n", i)
			return -1, -1, nil, true
		}
	}

	t.log.Printf("Insert: MAX %v evictions\n", MAX_EVICTIONS)
	return currBucketLoc.Bucket1, currBucketLoc.Bucket2, currVal, false
}

// Removes the entry from the cuckoo table
// Returns true if a value was removed, false if not
func (t *Table) Remove(bucket1 int, bucket2 int, value Comparable) bool {
	result := false
	result = result || t.removeFromBucket(bucket1, value)
	result = result || t.removeFromBucket(bucket2, value)
	return result
}

/********************
 * PRIVATE METHODS
 ********************/

// Checks if the `value` is in a specified bucket
// Returns true if `value.Equals(...)` returns true
func (t *Table) isInBucket(bucketIndex int, value Comparable) bool {
	if bucketIndex >= t.numBuckets {
		return false
	}
	bucket := t.buckets[bucketIndex]
	for i := 0; i < t.depth; i++ {
		if bucket.filled[i] && value.Compare(bucket.data[i]) == 0 {
			return true
		}
	}
	return false
}

// Tries to inserts `bucketLoc, value` into specified bucket
// If the bucket is already full, skip
// Returns true if success, false if bucket already full
func (t *Table) tryInsertToBucket(bucketIndex int, bucketLoc BucketLocation, value Comparable) bool {
	// Search for an empty slot
	bucket := t.buckets[bucketIndex]
	for i := 0; i < t.depth; i++ {
		if !bucket.filled[i] {
			bucket.data[i] = value
			bucket.bucketLoc[i] = bucketLoc
			bucket.filled[i] = true
			return true
		}
	}

	return false
}

// Tries to insert `bucketLoc, value` into specified bucket
// Returns -1, BucketLocation{}, nil, true if there's empty space and succeeds
// Returns false if insertion triggered an eviction
//   other values contain the evicted item's alternate bucket, BucketLocation pair, and value
func (t *Table) insertAndEvict(bucketIndex int, bucketLoc BucketLocation, value Comparable) (int, BucketLocation, Comparable, bool) {
	ok := t.tryInsertToBucket(bucketIndex, bucketLoc, value)
	if ok {
		return -1, BucketLocation{}, nil, true
	}

	// Eviction
	randIndex := t.rand.Int() % t.depth
	if randIndex < 0 {
		randIndex *= -1
	}
	bucket := t.buckets[bucketIndex]
	evictedBucketLoc := bucket.bucketLoc[randIndex]
	evictedValue := bucket.data[randIndex]
	bucket.bucketLoc[randIndex] = bucketLoc
	bucket.data[randIndex] = value
	bucket.filled[randIndex] = true
	if bucketIndex == evictedBucketLoc.Bucket1 {
		return evictedBucketLoc.Bucket2, evictedBucketLoc, evictedValue, false
	} else if bucketIndex == evictedBucketLoc.Bucket2 {
		return evictedBucketLoc.Bucket1, evictedBucketLoc, evictedValue, false
	} else {
		t.log.Fatalf("insertAndEvict: misplaced value! bucketIndex=%v does not match %v", bucketIndex, evictedBucketLoc)
		return -1, BucketLocation{}, nil, false
	}
}

// Removes all copies of `value` from the specified bucket
func (t *Table) removeFromBucket(bucketIndex int, value Comparable) bool {
	if bucketIndex >= t.numBuckets {
		return false
	}
	result := false
	bucket := t.buckets[bucketIndex]
	for i := 0; i < t.depth; i++ {
		if bucket.filled[i] && value.Compare(bucket.data[i]) == 0 {
			bucket.filled[i] = false
			bucket.data = nil
			bucket.bucketLoc[i] = BucketLocation{}
			result = true
		}
	}
	return result
}
