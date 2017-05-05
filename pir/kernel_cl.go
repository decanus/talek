package pir

// https://www.khronos.org/files/opencl-quick-reference-card.pdf

// db: shard ([bucket0, bucket1, ...]) where each bucket is bucketSize bytes
// reqs: batch of request vectors ([req0, req1, ...]) where each req is reqLength bytes
// output: batch of responses ([resp0, resp1, ...]) where each resp is bucketSize bytes
// scratch: L2 scratchpad of GPU_SCRATCH_SIZE bytes
// batchSize: number of requests per batch
// reqLength: length of a request in bytes (numBuckets/8)
// numBuckets: number of buckets in the shard
// bucketSize: length of a bucket in units of DATA_TYPE
// globalSize: number of threads globally (size of db if Kernel0, size of output if Kernel1)
// scratchSize: length of scratch in units of DATA_TYPE

const (
	GPU_SCRATCH_SIZE     = 2048 // Size of GPU scratch/L1 cache in bytes
	KERNEL_DATATYPE_SIZE = 8    // See DATA_TYPE in the kernel
)

// Workgroup == 1 request
// Workgroup items split up the scan over the database
const KernelCL0 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {

  int workgroup_size = get_local_size(0);
  int workgroup_index = get_local_id(0);
  int workgroup_num = get_group_id(0);
  int mask_offset = workgroup_num * (globalSize / bucketSize) / 8;
  long current_mask;
  short bitshift;

  // zero scratch
  for (int offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
    scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  for (int offset = workgroup_index; offset < globalSize; offset += workgroup_size) {
    bitshift = offset / bucketSize % 8;
    current_mask = reqs[mask_offset + offset / bucketSize / 8] & (1 << bitshift);
    current_mask = (current_mask >> bitshift) * -1;
    scratch[offset % bucketSize] ^= current_mask & db[offset];
  }

  // send to output.
  barrier(CLK_LOCAL_MEM_FENCE);
  for (int offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
    output[workgroup_num * bucketSize + offset] = scratch[offset];
  }
}
` + "\x00"

// index => output
// Cache the request
const KernelCL1 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
        __global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  int localSize = get_local_size(0);
  int localIndex = get_local_id(0);
  int groupIndex = get_group_id(0);
  int globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  DATA_TYPE result = 0;
  int reqIndex = (globalIndex / bucketSize) * reqLength;
  int offset = globalIndex % bucketSize;
  unsigned char reqBit;
  for (int i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;

}
` + "\x00"

// index => output
// Cache a portion of the database
const KernelCL2 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  int localSize = get_local_size(0);
  int localIndex = get_local_id(0);
  int groupIndex = get_group_id(0);
  int globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  DATA_TYPE result = 0;
  int reqIndex = (globalIndex / bucketSize) * reqLength;
  int offset = globalIndex % bucketSize;
  unsigned char reqBit;
  for (int i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;

}
` + "\x00"

// index => db
// Cache portion of the database
const KernelCL3 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  //int groupIndex = get_group_id(0);
  //int localSize = get_local_size(0);
  int globalIndex = get_global_id(0);
  int localIndex = get_local_id(0);

  if (globalIndex < globalSize) {
    // Zero output
    int outputSize = batchSize * bucketSize;
    if (globalSize >= outputSize && globalIndex < outputSize) {
      output[globalIndex] = 0;
    } else if (globalSize < outputSize) {
      int multiplier = outputSize / globalSize + 1;
      int start = globalIndex * multiplier;
      int end = start + multiplier;
      for (int i = start; i < end && i < outputSize; i++) {
	output[i] = 0;
      }
    }
    barrier(CLK_GLOBAL_MEM_FENCE);

    // Iterate over a batch
    for (int i = 0; i < batchSize; i++) {
      if () {
      }
    }

  }
}
` + "\x00"
