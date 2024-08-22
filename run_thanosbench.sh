#!/bin/bash

# Variables
range="$1"

# Parse the range
IFS=',' read -r start end <<< "$range"
start=$(printf "%d" "$start")
end=$(printf "%d" "$end")


# Constants
NUM_NAMESPACES=200
MAX_TIME="2024-06-07T00:00:00Z"

# Function to generate a random number in the given range
random_in_range() {
  local min=$1
  local max=$2
  echo $(awk -v min=$min -v max=$max 'BEGIN{srand(); print min + rand() * (max - min)}')
}

# Loop through clusters and namespaces to generate blocks
for ((cluster = start; cluster <= end; cluster++)); do
  for namespace in $(seq 1 $NUM_NAMESPACES); do
    PROFILE="cc-1w-small-rs"
    
    MIN_GAUGE=$(random_in_range 2.1 4.6)
    MAX_GAUGE=$(random_in_range 10.6 19.8)
    
    # Output directory based on weeks
    OUTPUT_DIR="./clusters-300/cluster-${cluster}"
    mkdir -p $OUTPUT_DIR
    
    # Generate the block plan and blocks
    MIN_GAUGE=$MIN_GAUGE MAX_GAUGE=$MAX_GAUGE ./thanosbench block plan -p $PROFILE \
      --labels "namespace=\"bench-${namespace}\"" \
      --labels "instance=\"bench\"" \
      --labels "cluster=\"ac-test-man-${cluster}\"" \
      --labels "container=\"bench\"" \
      --labels "resource=\"cpu\"" \
      --labels "clusterType=\"bench\"" \
      --labels "mode=\"idle\"" \
      --labels "profile=\"Max OverAll\"" \
      --max-time $MAX_TIME \
      | ./thanosbench block gen --output.dir $OUTPUT_DIR --workers 20
  done
done

echo "Block generation completed for clusters from $start to $end and $NUM_NAMESPACES namespaces."