#!/usr/bin/env bash

cp -f deploy/crds/clickhouse.service.diamond.sensetime.com_clickhouseclusters_crd.yaml install/shell/clickhouseclusters_crd.yaml
cp -f deploy/crds/clickhouse.service.diamond.sensetime.com_clickhouseclusters_crd.yaml install/helm/clickhouse-operator/template/clickhouseclusters_crd.yaml

tag=$(git tag -l | tail -n 1)
sed -i "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-operator/values.yaml
sed -i "s/tag:.*/tag: ${tag}/g" install/helm/clickhouse-broker/values.yaml