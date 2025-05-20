#!/bin/sh

KAFKA_BROKER="redpanda:29092"
KAFKA_TOPIC=${1:-"events"}

for file in /kafka_events/events/*; do
  echo "Publishing event: $file to topic $KAFKA_TOPIC using broker $KAFKA_BROKER"
  kcat -b $KAFKA_BROKER -t $KAFKA_TOPIC -P $file
done
