#!/bin/env bash
set -e

CWD=$(cd "$(dirname $0)"; pwd)
UTIL_SH="${CWD}/../util.sh"

source "${UTIL_SH}"

clickhouse_cluster="${1:?'miss chc name'}"
namespace="${2:?'miss namespace'}"

wait_for_chc_ready "${clickhouse_cluster}" "${namespace}"

statefulsets=$(get_statefulsets_from_chc "${clickhouse_cluster}" "${namespace}")

for statefulset in ${statefulsets};
do
  ready_num=$(kubectl get statefulset "${statefulset}" --namespace "${namespace}" -o jsonpath='{.status.readyReplicas}')

  for((i=0;i<ready_num;i++));
  do
    pod_name="${statefulset}"-$i
    host="$pod_name"."${statefulset}"."${namespace}".svc.cluster.local

    query="create database if not exists test"
    kubectl exec "$pod_name" --namespace "${namespace}" -c clickhouse -- clickhouse-client -h "$host" --query="${query}";

    query=$(cat "${CWD}"/table.sql)
    kubectl exec "$pod_name" --namespace "${namespace}" -c clickhouse -- clickhouse-client -h "$host" -d test --query="${query}";

    query="insert into test_table(CounterID, UserID) values($i, 9527)"
    kubectl exec "$pod_name" --namespace "${namespace}" -c clickhouse -- clickhouse-client -h "$host" -d test --query="${query}";

    sleep 1s
    kubectl delete pod "$pod_name" --namespace "${namespace}" --force --grace-peirod=0
  done
done

wait_for_chc_ready "${clickhouse_cluster}" "${namespace}"

query="select CounterID from test_table where UserID=9527"
for statefulset in ${statefulsets};
do
  pod_name="${statefulset}"-$i
  host="$pod_name"."${statefulset}"."${namespace}".svc.cluster.local
  counter_id=$(kubectl exec "$pod_name" --namespace "${namespace}" -c clickhouse -- clickhouse-client -h "$host" -d test --query="${query}");
  if [ "${counter_id}" != "$i" ];then
    echo "counter_id is not $i"
    exit 1
  fi
done
