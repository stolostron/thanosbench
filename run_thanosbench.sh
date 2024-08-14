#!/bin/bash

# Variables
NUM_CLUSTERS=300
NUM_NAMESPACES=200
MAX_TIMES=("2024-06-07T00:00:00Z" "2024-06-14T00:00:00Z" "2024-06-21T00:00:00Z" "2024-06-28T00:00:00Z" "2024-07-05T00:00:00Z")

random_in_range() {
  local min=$1
  local max=$2
  echo $(awk -v min=$min -v max=$max 'BEGIN{srand(); print min + rand() * (max - min)}')
}


# loop through clusters and namespaces to generate blocks
for cluster in $(seq 1 $NUM_CLUSTERS); do
  for namespace in $(seq 1 $NUM_NAMESPACES); do
    for i in ${!MAX_TIMES[@]}; do
      PROFILE="cc-1w-small-rs"
      MAX_TIME=${MAX_TIMES[$i]}
  
      MIN_GAUGE=$(random_in_range 2.1 4.6)
      MAX_GAUGE=$(random_in_range 10.6 19.8)

      # output directory based on weeks
      OUTPUT_DIR="./month-1/week-$((i + 1))"

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
done

echo "Block generation completed for $NUM_CLUSTERS clusters and $NUM_NAMESPACES namespaces."
