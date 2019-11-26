#!/bin/env bash
set -e

STATEFULSET="${1:?'miss statefulset name'}"
NAMESPACE="${2:?'miss namespace'}"

query="
CREATE TABLE test_table
(
    EventDate DateTime,
    CounterID UInt32,
    UserID UInt32
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/0-0/test_table', 'replica')
PARTITION BY toYYYYMM(EventDate)
ORDER BY (CounterID, EventDate, intHash32(UserID))
SAMPLE BY intHash32(UserID)"

ready_num=$(kubectl get statefulset "$STATEFULSET" --namespace "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')

for((i=0;i<${ready_num};i++));
do
  pod_name="$STATEFULSET"-$i
  host="$pod_name"."$STATEFULSET"."$NAMESPACE".svc.cluster.local
  kubectl exec "$pod_name" --namespace $NAMESPACE -- clickhouse-client -h "$host" --query="create database if not exists test";
  q=${query//replica/$pod_name}
  kubectl exec "$pod_name" --namespace $NAMESPACE -- clickhouse-client -h "$host" --query="${q}";
done
