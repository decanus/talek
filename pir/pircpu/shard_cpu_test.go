package pircpu

import (
	"fmt"
	pt "github.com/privacylab/talek/pir/pirtest"
	"testing"
)

func TestShardCPUCreate(t *testing.T) {
	fmt.Printf("TestShardCPUCreate: ...\n")
	// Creating with invalid bucketSize
	shard, err := NewShardCPU("shardcpuv0", 7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 0)
	if err == nil {
		t.Fatalf("new ShardCPU should have failed with invalid bucketSize, but didn't return error")
	}
	if shard != nil {
		t.Fatalf("new ShardCPU should have failed with invalid bucketSize, but returned a shard")
	}
	fmt.Printf("... done \n")
}

func TestShardCPUReadv0(t *testing.T) {
	fmt.Printf("TestShardCPUReadv0: ...\n")
	shard, err := NewShardCPU("shardcpuv0", pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv1(t *testing.T) {
	fmt.Printf("TestShardCPUReadv1: ...\n")
	shard, err := NewShardCPU("shardcpuv1", pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv2(t *testing.T) {
	fmt.Printf("TestShardCPUReadv2: ...\n")
	shard, err := NewShardCPU("shardcpuv2", pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 2)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v2: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func BenchmarkShardCPUReadv0(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv0", pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv1(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv1", pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), 1)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv2(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv2", pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), 2)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v2: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}
