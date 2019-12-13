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
  image: registry.sensetime.com/diamond/service-providers/clickhouse:latest
  initImage: registry.sensetime.com/diamond/clickhouse-bootstrap:latest
  shardsCount: 1
  replicasCount: 2
```

Each field is described as follows:

| Field              |                                 Describe                                 |
| ------------------ | :----------------------------------------------------------------------: |
| `image`            |              Base image to use for a Clickhouse deployment               |
| `initImage`        |                   Image used for bootstrapping cluster                   |
| `dataCapacity`     |  Define the Capacity for Persistent Volume Claims in the local storage   |
| `dataStorageClass` |  Define StorageClass for Persistent Volume Claims in the local storage   |
| `deletePVC`        | DeletePVC defines if the PVC must be deleted when the cluster is deleted |
| `shardsCount`      |                               Shards count                               |
| `replicasCount`    |                        clickhouse Replicas count                         |
| `zookeeper`        |                             Zookeeper config                             |
| `users`            |                              Users defined                               |
| `pod`              |                                POD config                                |
| `resources`        |      Pod defines the policy for pods owned by clickhouse operator.       |

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

More examples can be find in [samples](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/tree/master/samples)
