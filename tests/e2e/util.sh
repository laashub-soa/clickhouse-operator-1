#!/bin/env bash

function wait_for_chc_ready() {
  name=${1:?"lack clickhouse name"}
  namespace=${2:?"lack clickhouse namespace"}
  retry_num=0
  while true; do
    ready=$(kubectl get clickhousecluster "${name}" -n "${namespace}" -o jsonpath="{.status.phase}")
    if [ "${ready}" = "Running" ];then
      break
    fi
    if [ ${retry_num} -gt 10 ];then
        echo "the clickhousecluster is not ready"
        exit 1
    fi
    retry_num=$((retry_num+1))
    sleep 5s
  done
}

function get_statefulsets_from_chc() {
    name=${1:?"lack clickhouse name"}
    kubectl get statefulset -l clickhouse-cluster="${name}" -o jsonpath="{.items[*].metadata.name}"
}

function get_pods_from_statefulset() {
    name=${1:?"lack statefulset name"}
    kubectl get pod -l statefulset="${name}" -o jsonpath="{.items[*].metadata.name}"
}
