---
title: "Deploy clickhouse by svcat"
weight: 5
description: "clickhouse"
date: 2019-8-022T9:30:00+08:00
---

### Create a service instance

Firstly, you should check out the available service plans. All parameters the plan includes can be modified:

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

We can easily use plan `default` to provision a 2-nodes Clickhouse cluster:

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

Checkout the instance's status:

```bash
$ svcat get instances --all-namespaces
  NAME   NAMESPACE                  CLASS                                   PLAN                   STATUS
+------+-----------+--------------------------------------+--------------------------------------+--------+
  demo   test        137a3ded-59ab-4ece-bbda-9cfff850a1f3   2db0b31d-6912-4d24-8704-cfdf9b98af81   Ready
```

Checkout the status of pods:

```bash
$ kc get po -n test
NAME                     READY   STATUS    RESTARTS   AGE
clickhouse-my-cluster-shard-0   1/1     Running   0          100s
```

### Bind the instance

Use binding command:

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

View the binding status:

```bash
$ svcat get bindings --all-namespaces
  NAME   NAMESPACE   INSTANCE   STATUS
+------+-----------+----------+--------+
  demo   test        demo       Ready
```

Due to the startup time of the cluster, we recommend you making a bind 3 minutes later after provisioning a cluster. Otherwise, the binding could fail. If the binding status was `Ready`, a secret named `demo` would be newed automatically in this binding, which contains username and password of the cluster and can be directly utilized by clients.

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

Username and password here are randomly generated. You can also assign your username and password by using the command below:

```bash
svcat bind demo -n test -p user=mike -p password=123456
```

### Checkout Clickhouse cluster status

Login Grafana, you can see related statistics about the cluster:
![](/services/img/clickhouse/grafana.png)

### Unbind

```bash
$ svcat unbind demo -n test
deleted demo
```

The generated/assigned username and password will be invalid immediately.

### Delete the instance

```bash
$ svcat deprovision demo -n test
deleted demo
```
