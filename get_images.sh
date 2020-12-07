#!/bin/env bash

CWD=$(cd `dirname $0`; pwd)

OperatorImage=$(cat ${CWD}/install/helm/clickhouse-operator/values.yaml |grep 'repository:' | awk '{print $2}')
OperatorTag=$(cat ${CWD}/install/helm/clickhouse-operator/values.yaml |grep 'tag:' | awk '{print $2}')

if [ -e $OperatorImage -o -e $OperatorTag ];then
  echo "find operator image error" >&2
  exit 1
fi

echo "$OperatorImage":"$OperatorTag"


ZKImage=$(cat ${CWD}/install/zookeeper/zookeeper.yaml |grep -E 'image:' | awk '{print $2}' | sed 's/\"//g')
if [ -e "$ZKImage" ];then
  echo "find zk image error" >&2
  exit 1
fi

echo "$ZKImage"

OtherImage=$(cat ${CWD}/install/helm/clickhouse-operator/templates/operator-configmap.yaml \
|grep -E 'default_clickhouse_image:|default_clickhouse_exporter_image:' | awk '{print $2}' \
| sed 's/\"//g')

if [ -e "$OtherImage" ];then
  echo "find other images error" >&2
  exit 1
fi

echo "$OtherImage"
