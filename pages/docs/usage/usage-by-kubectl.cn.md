---
title: "通过ClickhouseCluster部署"
weight: 4
description: "clickhouse"
date: 2019-7-018T16:36:04+08:00
---

这是一个 ClickhouseCluster 例子：

```yaml
#clickhousecluster.yaml
apiVersion: clickhouse.service.diamond.sensetime.com/v1
kind: ClickHouseCluster
metadata:
  name: simple
  namespace: test
spec:
  shardsCount: 1
  replicasCount: 2
```

每个字段说明如下：

| 字段               |           含义           |
| ------------------ | :----------------------: |
| `image`            |    clickhouse 镜像名     |
| `initImage`        |   用于初始化群集的镜像   |
| `dataCapacity`     |  Storage Class 容量大小  |
| `dataStorageClass` |    Storage Class 名称    |
| `deletePVC`        |  删除集群时是否删除 pvc  |
| `shardsCount`      |  clickhouse shard 数量   |
| `replicasCount`    | clickhouse replicas 数量 |
| `zookeeper`        |      zookeeper 配置      |
| `users`            |         用户配置         |
| `pod`              |         POD 配置         |
| `resources`        |         资源配置         |

创建/更新实例

```bash
$ kubectl apply -f clickhousecluster.yaml -n clickhouse-namepace
clickhousecluster.sensetime.com/clickhouse-demo created
```

查看 pod 状态：

```bash
$ kubectl get po -n clickhouse-namepace
NAME                         READY   STATUS    RESTARTS   AGE
clickhouse-demo-dc1-rack1-0   1/1     Running   0          36m
clickhouse-demo-dc1-rack2-0   1/1     Running   0          35m
```

更多实例请参考 [samples](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/tree/master/samples)
