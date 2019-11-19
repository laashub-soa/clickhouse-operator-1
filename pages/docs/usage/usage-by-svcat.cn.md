---
title: "通过svcat部署"
weight: 4
description: "clickhouse"
date: 2019-7-018T16:36:04+08:00
---

### 创一个 service instance

1. 首先查看有哪些可用的 plan, plan 所列的参数的值都支持自定义。

```bash
$ svcat get plan -c clickhouse -o yaml
- metadata:
    name: 2db0b31d-6912-4d24-8704-cfdf9b98af81
  spec:
    clusterServiceBrokerName: clickhouse-service-broker
    clusterServiceClassRef:
      name: 137a3ded-59ab-4ece-bbda-9cfff850a1f3
    description: The default plan for the clickhouse service
    externalID: 2db0b31d-6912-4d24-8704-cfdf9b98af81
    externalName: default
    free: true
    instanceCreateParameterSchema:
      dataCapacity: 20Gi
      deletePVC: true
      shardCount: 1
```

我们可以通过默认套餐快速创建一个 2 节点的 Clickhouse 集群:

```bash
$ svcat provision demo --class clickhouse --plan default --param cluster_name=my-cluster -p dataStorageClass=ceph-rbd --namespace test
  Name:        demo
  Namespace:   test
  Status:
  Class:       137a3ded-59ab-4ece-bbda-9cfff850a1f3
  Plan:        2db0b31d-6912-4d24-8704-cfdf9b98af81

Parameters:
  cluster_name: my-cluster
  dataStorageClass: ceph-rbd
```

查看 instance 状态：

```bash
$ svcat get instances --all-namespaces
  NAME   NAMESPACE                  CLASS                                   PLAN                   STATUS
+------+-----------+--------------------------------------+--------------------------------------+--------+
  demo   test        137a3ded-59ab-4ece-bbda-9cfff850a1f3   2db0b31d-6912-4d24-8704-cfdf9b98af81   Ready
```

查看 pod 状态：

```bash
$ kc get po -n test
NAME                     READY   STATUS    RESTARTS   AGE
clickhouse-my-cluster-shard-0   1/1     Running   0          100s
```

### 绑定一个实例

执行绑定命令：

```bash
$ svcat bind demo -n test
  Name:        demo
  Namespace:   test
  Status:
  Secret:      demo
  Instance:    demo

Parameters:
  No parameters defined
```

查看绑定状态：

```bash
$ svcat get bindings --all-namespaces
  NAME   NAMESPACE   INSTANCE   STATUS
+------+-----------+----------+--------+
  demo   test        demo       Ready
```

由于集群启动需要时间，建议在创建集群后 3 分钟后进行绑定操作，否则可能会出绑定失败的问题。
如果绑定状态显示 `Ready`，则会在 `demo`中自动生成一个名为`demo`的 secret，里面保存有用户名密码,客户端可以直接挂载使用。

```bash
$ svcat describe binding demo -n test --show-secrets
apiVersion: v1
data:
  host: WyJteS1jbHVzdGVyLWRjMS1yYWNrMS0wLm15LWNsdXN0ZXIudGVzdCIsIm15LWNsdXN0ZXItZGMxLXJhY2sxLTEubXktY2x1c3Rlci50ZXN0Il0=
  password: dUNFZXlrSmlidXpCUVl6Vkhib3I=
  user: WldSa3Y=
kind: Secret
metadata:
  name: demo
  namespace: test
type: Opaque
```

用户名和密码都是随机生成，如果要指定用户名和密码，可以使用：

```bash
svcat bind demo -n test -p user=mike -p password=123456
```

### 查看 Clickhouse 集群状态

登录 grafana， 可以查看集群各种相关参数：
![](/services/img/clickhouse/grafana.png)

### 解除绑定

```bash
$ svcat unbind demo -n test
deleted demo
```

解除绑定后，该绑定所生成的用户名和密码也会立即失效。

### 删除实例

```bash
$ svcat deprovision demo -n test
deleted demo
```
