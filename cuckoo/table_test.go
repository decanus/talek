package cuckoo

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

const ItemSize = 64

// Test identical state after operations
// Test contains after insert/remove sequence
// Test insert same value twice

func GetBytes(val string) []byte {
	buf := make([]byte, ItemSize)
	copy(buf, bytes.NewBufferString(val).Bytes())
	return buf
}

func randBucket(numBuckets int) int {
	result := rand.Int() % numBuckets
	if result < 0 {
		result = result * -1
	}
	return result
}

func TestGetCapacity(t *testing.T) {
	fmt.Printf("TestGetCapacity: ...\n")

	table := NewTable("t", 10, 2, ItemSize, nil, 0)
	if table.GetCapacity() != 20 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 20\n", table.GetCapacity())
	}

	table = NewTable("t", 1, 1, ItemSize, nil, 0)
	if table.GetCapacity() != 1 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 1\n", table.GetCapacity())
	}

	table = NewTable("t", 0, 0, ItemSize, nil, 0)
	if table.GetCapacity() != 0 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 0\n", table.GetCapacity())
	}

	fmt.Printf("... done \n")
}

func TestInvalidConstruction(t *testing.T) {
	fmt.Printf("TestInvalidConstruction: ...\n")
	data := make([]byte, 7)
	table := NewTable("t", 2, 3, 4, data, 0)
	if table != nil {
		t.Fatalf("created a improperly sized table\n")
	}
	fmt.Printf("... done \n")
}

func TestBasic(t *testing.T) {
	table := NewTable("t", 10, 2, ItemSize, nil, 0)

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Check contains non-existent value...\n")
	if table.Contains(&Item{0, GetBytes(""), 0, 1}) {
		t.Fatalf("empty table returned true for Contains()\n")
	}

	fmt.Printf("TestBasic: remove non-existent value ...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Insert improperly sized value ...\n")
	ok, itm := table.Insert(&Item{1, []byte{0, 0}, 0, 1})
	if itm != nil || ok {
		t.Fatalf("should have failed inserting a malformed item\n")
	}

	fmt.Printf("TestBasic: Insert value ...\n")
	ok, itm = table.Insert(&Item{1, GetBytes("value1"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("error inserting into table (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: Check inserted value...\n")
	if !table.Contains(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("cannot find recently inserted value\n")
	}

	fmt.Printf("TestBasic: Check inserted value w/o full reference...\n")
	if !table.Contains(&Item{1, GetBytes(""), 0, 1}) {
		t.Fatalf("cannot find recently inserted value\n")
	}

	fmt.Printf("TestBasic: Check non-existent value...\n")
	if table.Contains(&Item{2, GetBytes("value2"), 0, 1}) {
		t.Fatalf("contains a non-existent value\n")
	}

	fmt.Printf("TestBasic: check 1 element...\n")
	if table.GetNumElements() != 1 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove existing value ...\n")
	if !table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("error removing existing value (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove recently removed value ...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("... done \n")
}

func TestBucket(t *testing.T) {
	fmt.Printf("TestBucket ...\n")
	table := NewTable("t", 10, 2, ItemSize, nil, 0)

	items := []*Item{
		{1, GetBytes("value1"), 5, 5},
		{2, GetBytes("value2"), 5, 5},
		{3, GetBytes("value3"), 5, 6},
	}
	for _, v := range items {
		ok, _ := table.Insert(v)
		if !ok {
			t.Fatalf("Failed to insert item %v", v)
		}
	}

	if table.Bucket(items[0]) != 5 {
		t.Fatalf("Table should report expected item position %v", items[0])
	}
	if table.Bucket(items[1]) != 5 {
		t.Fatalf("Table should report expected item position %v", items[1])
	}
	if table.Bucket(items[2]) != 6 {
		t.Fatalf("Table should report expected item position %v", items[2])
	}

	nonItem := &Item{4, GetBytes("value4"), 1, 1}
	if table.Bucket(nonItem) != -1 {
		t.Fatalf("Table should report expected item position %v", nonItem)
	}

	fmt.Printf("... done \n")
}

func TestOutOfBounds(t *testing.T) {
	table := NewTable("t", 10, 2, ItemSize, nil, 0)

	fmt.Printf("TestOutOfBounds: Insert() out of bounds...\n")
	ok, itm := table.Insert(&Item{1, GetBytes("value1"), 100, 100})
	if ok {
		t.Fatalf("Insert returned true with out of bound buckets\n")
	}
	if itm != nil {
		t.Fatalf("Insert returned wrong values\n")
	}

	fmt.Printf("TestOutOfBounds: Contains() out of bounds...\n")
	if table.Contains(&Item{1, GetBytes("value1"), 100, 100}) {
		t.Fatalf("Contains returned true with out of bound buckets\n")
	}

	fmt.Printf("TestOutOfBounds: Remove() out of bounds...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 100, 100}) {
		t.Fatalf("Remove() returned true with out of bound buckets\n")
	}

	fmt.Printf("... done \n")
}

func TestFullTable(t *testing.T) {
	numBuckets := 100
	depth := 4

	capacity := numBuckets * depth
	entries := make([]Item, 0, capacity)
	table := NewTable("t", numBuckets, depth, ItemSize, nil, 0)
	ok := true
	count := 0
	var evic *Item
	var b1, b2 int

	// Insert random values until we've reached a limit
	fmt.Printf("TestFullTable: Insert until reach capacity...\n")
	for ok {
		b1 = randBucket(numBuckets)
		b2 = randBucket(numBuckets)
		id := rand.Int()
		val := GetBytes(strconv.Itoa(rand.Int()))
		entries = append(entries, Item{id, nil, b1, b2})
		ok, evic = table.Insert(&Item{id, val, b1, b2})

		if ok {
			count++
			if !table.Contains(&Item{id, nil, b1, b2}) {
				t.Fatalf("Insert() succeeded, but Contains failed\n")
			}
			if count != table.GetNumElements() {
				t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n",
					count,
					table.GetNumElements())
			}
		}
	}

	// Middle count check
	fmt.Printf("TestFullTable: Fully Loaded check...\n")
	if count != table.GetNumElements() {
		t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n",
			count,
			table.GetNumElements())
	}
	maxCount := count

	// Remove elements one by one
	fmt.Printf("TestFullTable: Remove each element...\n")
	for _, entry := range entries {
		if !entry.Equals(evic) {
			ok = table.Remove(&entry)
			if !ok {
				t.Fatalf("Cannot Remove() an element believed to be in the table. item %d of %d",
					count,
					maxCount)
			} else {
				count--
				if count != table.GetNumElements() {
					t.Fatalf("GetNumElements()=%v returned a value that didn't match what was expected=%v \n",
						table.GetNumElements(),
						count)
				}
			}
		}
	}

	// Final count check
	fmt.Printf("TestFullTable: Final count check...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("GetNumElements() returns %v when table should be empty \n", table.GetNumElements())
	}

	fmt.Printf("... done\n")
}

func TestDuplicateValues(t *testing.T) {
	fmt.Printf("TestDuplicateValues: ...\n")
	table := NewTable("t", 10, 2, ItemSize, nil, 0)

	ok, itm := table.Insert(&Item{1, GetBytes("v"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value \n")
	}

	ok, itm = table.Insert(&Item{2, GetBytes("v"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value again \n")
	}

	ok, itm = table.Insert(&Item{3, GetBytes("v"), 1, 2})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value in shifted buckets\n")
	}

	if !table.Remove(&Item{1, GetBytes("v"), 0, 1}) {
		t.Fatalf("Error removing value 1st time\n")
	}

	if !table.Remove(&Item{2, GetBytes("v"), 0, 1}) {
		t.Fatalf("Error removing value 2nd time\n")
	}

	if !table.Remove(&Item{3, GetBytes("v"), 1, 2}) {
		t.Fatalf("Error removing value 3rd time\n")
	}

	fmt.Printf("... done\n")
}

func TestLoadFactor(t *testing.T) {
	fmt.Printf("TestLoadFactor: ...\n")
	numBuckets := 1000
	var table *Table

	for depth := 1; depth < 32; depth *= 2 {
		count := -1
		ok := true
		table = NewTable("t", numBuckets, depth, ItemSize, nil, int64(depth))
		for ok {
			count++
			val := GetBytes(strconv.Itoa(rand.Int()))
			ok, _ = table.Insert(&Item{rand.Int(), val, randBucket(numBuckets), randBucket(numBuckets)})
		}

		if table.GetNumElements() != count {
			t.Fatalf("Mismatch count=%v with GetNumElements()=%v \n", count, table.GetNumElements())
		}
		fmt.Printf("count=%v, depth=%v, capacity=%v, loadfactor=%v \n",
			count,
			depth,
			table.GetCapacity(),
			(float64(count) / float64(table.GetCapacity())))
	}

	fmt.Printf("... done\n")
}
