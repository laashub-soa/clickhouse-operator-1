---
title: "Deploy clickhouse by ClickhouseCluster"
weight: 4
description: "clickhouse"
date: 2019-7-018T16:36:04+08:00
---

Here is an example of Clickhouse cluster：

```yaml
#clickhousecluster.yaml
apiVersion: clickhouse.service.diamond.sensetime.com/v1
kind: ClickHouseCluster
metadata:
  name: simple
  namespace: test
spec:
  baseImage: registry.sensetime.com/diamond/service-providers/clickhouse:latest
  bootstrapImage: registry.sensetime.com/diamond/clickhouse-bootstrap:latest
  shardsCount: 1
  replicasCount: 2
```

Each field is described as follows:

| Field              |                                 Describe                                 |
| ------------------ | :----------------------------------------------------------------------: |
| `baseImage`        |              Base image to use for a Clickhouse deployment               |
| `bootstrapImage`   |                   Image used for bootstrapping cluster                   |
| `dataCapacity`     |  Define the Capacity for Persistent Volume Claims in the local storage   |
| `dataStorageClass` |  Define StorageClass for Persistent Volume Claims in the local storage   |
| `deletePVC`        | DeletePVC defines if the PVC must be deleted when the cluster is deleted |

Create clickhouse instance

```bash
$ kc create -f clickhousecluster.yaml -n clickhouse-namepace
clickhousecluster.sensetime.com/clickhouse-demo created
```

Check Pod status：

```bash
$ kubectl get po -n clickhouse-namepace
NAME                         READY   STATUS    RESTARTS   AGE
clickhouse-demo-dc1-rack1-0   1/1     Running   0          36m
clickhouse-demo-dc1-rack2-0   1/1     Running   0          35m
```
