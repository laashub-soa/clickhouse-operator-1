---
title: "通过ClickhouseCluster部署"
weight: 4
description: "clickhouse"
date: 2019-7-018T16:36:04+08:00
---

这是一个 ClickhouseCluster 例子：

```yaml
#clickhousecluster.yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "ClickhouseCluster"
metadata:
  name: clickhouse-demo
spec:
  baseImage: registry.sensetime.com/diamond/service-providers/clickhouse
  version: 3.11.4-8u212-0.3.2-release-cqlsh
  bootstrapImage: registry.sensetime.com/diamond/clickhouse-bootstrap:0.1.0
  # dataCapacity: "200Mi"
  # dataStorageClass: "local-storage"
  # pod:
  #   annotations:
  #     exemple.com/test: my.custom.annotation
  #   tolerations:
  #     - effect: NoExecute
  #       key: fack-tolerations
  #       operator: Exists
  #       tolerationSeconds: 300
  ImagePullPolicy: IfNotPresent
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  gcStdout: true
  autoUpdateSeedList: true
  maxPodUnavailable: 1
  runAsUser: 1000
  resources:
    requests:
      cpu: "1"
      memory: 2Gi
    limits:
      cpu: "1"
      memory: 2Gi
  topology:
    dc:
      - name: dc1
        nodesPerRacks: 1
        rack:
          - name: rack1
          - name: rack2
```

每个字段说明如下：

| 字段                           |                       含义                       |
| ------------------------------ | :----------------------------------------------: |
| `baseImage`                    |                clickhouse 镜像名                 |
| `version`                      |                    镜像版本号                    |
| `bootstrapImage`               |               用于初始化群集的镜像               |
| `dataCapacity`                 |              Storage Class 容量大小              |
| `dataStorageClass`             |                Storage Class 名称                |
| `pod`                          | pod 的相关信息，支持`annotations`、`tolerations` |
| `imagepullpolicy`              |                   镜像拉取策略                   |
| `hardAntiAffinity`             |   pod 的 antiAffinity 是否为 hard，默认为 soft   |
| `deletePVC`                    |              删除集群时是否删除 pvc              |
| `autoPilot`                    |               运维操作是否自动执行               |
| `gcStdout`                     |                 是否输出 GC 日志                 |
| `autoUpdateSeedList`           |     集群拓扑在变化时，是否自动更新 seed 列表     |
| `maxPodUnavailable`            |   Pod Disruption Budget 中`maxPodUnavailable`    |
| `runAsUser`                    |         clickhouse 容器中运行的用户的 id         |
| `resources`                    |               Clickhouse 资源需求                |
| `topology`                     |                 Clickhouse 拓扑                  |
| `topology.dc`                  |                     数据中心                     |
| `topology.dc[*].name`          |                   数据中心名称                   |
| `topology.dc[*].nodesPerRacks` |                 每个 rack 节点数                 |
| `topology.dc[*].rack`          |                    rack 信息                     |

创建/更新实例

```bash
$ kubectl apply -f clickhousecluster.yaml -n clickhouse-namepace
clickhousecluster.db.orange.com/clickhouse-demo created
```

查看 pod 状态：

```bash
$ kubectl get po -n clickhouse-namepace
NAME                         READY   STATUS    RESTARTS   AGE
clickhouse-demo-dc1-rack1-0   1/1     Running   0          36m
clickhouse-demo-dc1-rack2-0   1/1     Running   0          35m
```
