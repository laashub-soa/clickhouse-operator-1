#!/usr/bin/env bash
set -e

cp -f deploy/crds/clickhouse.service.diamond.sensetime.com_clickhouseclusters_crd.yaml install/helm/clickhouse-operator/templates/clickhouseclusters_crd.yaml

tag=$(git tag --sort=committerdate | tail -n 1)
os=`uname`

if [ ${os} = "Darwin" ];then
  sed -i "" "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-operator/values.yaml
  sed -i "" "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-broker/values.yaml
else
   sed -i "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-operator/values.yaml
   sed -i "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-broker/values.yaml
fi
