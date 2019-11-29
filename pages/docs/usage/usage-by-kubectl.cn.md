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
  baseImage: registry.sensetime.com/diamond/service-providers/clickhouse:latest
  bootstrapImage: registry.sensetime.com/diamond/clickhouse-bootstrap:latest
  shardsCount: 1
  replicasCount: 2
```

每个字段说明如下：

| 字段                           |                       含义                       |
| ------------------------------ | :----------------------------------------------: |
| `baseImage`                    |                clickhouse 镜像名                 |
| `bootstrapImage`               |               用于初始化群集的镜像               |
| `dataCapacity`                 |              Storage Class 容量大小              |
| `dataStorageClass`             |                Storage Class 名称                |
| `deletePVC`                    |              删除集群时是否删除 pvc              |

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
