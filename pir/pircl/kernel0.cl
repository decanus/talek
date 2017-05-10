typedef char int8_cl;
typedef unsigned char uint8_cl;
typedef int int32_cl;
typedef unsigned int uint32_cl;
typedef long int64_cl;
typedef unsigned long uint64_cl;

#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable
#define DATA_TYPE uint64_cl

__kernel
void pir(__global DATA_TYPE* db,
	__global uint8_cl* reqs,
  __global DATA_TYPE* output,
  __local DATA_TYPE* scratch,
  __const uint32_cl batchSize,
	__const uint32_cl reqLength,
	__const uint32_cl numBuckets,
	__const uint32_cl bucketSize,
	__const uint32_cl globalSize,
	__const uint32_cl scratchSize) {
  //uint32_cl globalSize = get_global_size(0);
  //uint32_cl localSize = get_local_size(0);
  //uint32_cl localIndex = get_local_id(0);
  //uint32_cl groupIndex = get_group_id(0);
  //uint32_cl globalIndex = get_global_id(0);
  uint32_cl workgroup_size = get_local_size(0);
  uint32_cl workgroup_index = get_local_id(0);
  uint32_cl workgroup_num = get_group_id(0);	  // request index

  // zero scratch
  for (uint32_cl offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
      scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  uint32_cl dbSize = numBuckets * bucketSize;
  uint32_cl reqIndex = workgroup_num * reqLength;
  uint32_cl bucketId;
  uint32_cl depthOffset;
  uint8_cl reqBit;
  for (uint32_cl offset = workgroup_index; offset < dbSize; offset += workgroup_size) {
    bucketId = offset / bucketSize;
    depthOffset = offset % bucketSize;
    reqBit = reqs[reqIndex + (bucketId/8)] & (1 << (bucketId%8));
    //current_mask = (current_mask >> bitshift) * -1;
    //scratch[depthOffset] ^= current_mask & db[offset];
    if (reqBit != 0) {
      //scratch[depthOffset] ^= db[offset];
      atom_xor(&scratch[depthOffset], db[offset]);
    }
  }

  // send to output.
  barrier(CLK_LOCAL_MEM_FENCE);
  uint32_cl respIndex = workgroup_num * bucketSize;
  for (uint32_cl offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
    output[respIndex + offset] = scratch[offset];
  }

}
